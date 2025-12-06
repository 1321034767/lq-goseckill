package service

import (
	"context"
	"time"

	"github.com/example/goseckill/internal/datamodels/product"
	"github.com/example/goseckill/internal/datamodels/seckill_activity"
)

// SeckillActivityService 秒杀活动领域服务
// 负责：
//   - 活动的创建 / 更新 / 删除
//   - 活动与商品的关联维护
//   - 根据时间窗口自动更新活动状态
//   - 启动活动时同步商品状态与秒杀库存到 Redis
//   - 为前台/后台提供活动查询能力

type SeckillActivityService struct {
	activityRepo seckill_activity.Repository
	productRepo  product.Repository
}

// NewSeckillActivityService 创建秒杀活动服务
func NewSeckillActivityService(activityRepo seckill_activity.Repository, productRepo product.Repository) *SeckillActivityService {
	return &SeckillActivityService{
		activityRepo: activityRepo,
		productRepo:  productRepo,
	}
}

// CreateActivity 创建秒杀活动
func (s *SeckillActivityService) CreateActivity(ctx context.Context, req *CreateActivityRequest) (*seckill_activity.SeckillActivity, error) {
	activity := &seckill_activity.SeckillActivity{
		Name:         req.Name,
		Description:  req.Description,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Discount:     req.Discount,
		LimitPerUser: req.LimitPerUser,
		Status:       0, // 默认未开始
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		return nil, err
	}

	// 维护活动与商品的关联及各自的秒杀库存
	for _, productID := range req.ProductIDs {
		p, err := s.productRepo.GetByID(ctx, productID)
		if err != nil {
			// 商品不存在则跳过
			continue
		}
		stock := req.ProductStocks[productID]
		if stock <= 0 {
			// 未显式指定则默认使用商品当前库存
			stock = p.Stock
		}
		// 秒杀库存从商品现有库存中划拨
		if stock > p.Stock {
			stock = p.Stock
		}
		p.Stock = p.Stock - stock
		if err := s.productRepo.Update(ctx, p); err != nil {
			// 划拨失败跳过该商品，避免影响其它商品
			continue
		}
		if err := s.activityRepo.AddProduct(ctx, activity.ID, productID, stock); err != nil {
			// 单个商品失败不影响整体
			continue
		}
	}

	return activity, nil
}

// UpdateActivity 更新活动基础信息（不包含商品列表）
func (s *SeckillActivityService) UpdateActivity(ctx context.Context, id int64, req *UpdateActivityRequest) error {
	activity, err := s.activityRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	activity.Name = req.Name
	activity.Description = req.Description
	activity.StartTime = req.StartTime
	activity.EndTime = req.EndTime
	activity.Discount = req.Discount
	if req.LimitPerUser > 0 {
		activity.LimitPerUser = req.LimitPerUser
	}

	return s.activityRepo.Update(ctx, activity)
}

// UpdateActivityProducts 重新配置某个活动下的商品及其秒杀库存
func (s *SeckillActivityService) UpdateActivityProducts(ctx context.Context, activityID int64, productIDs []int64, productStocks map[int64]int64) error {
	// 先读取当前关联关系
	existing, err := s.activityRepo.GetProductsByActivity(ctx, activityID)
	if err != nil {
		return err
	}

	existingMap := make(map[int64]bool)
	for _, ep := range existing {
		existingMap[ep.ProductID] = true
	}

	newMap := make(map[int64]bool)
	for _, id := range productIDs {
		newMap[id] = true
	}

	// 删除不再需要的商品
	for _, ep := range existing {
		if !newMap[ep.ProductID] {
			// 归还此前划拨的秒杀库存到商品库存
			if p, err := s.productRepo.GetByID(ctx, ep.ProductID); err == nil {
				p.Stock += ep.SeckillStock
				_ = s.productRepo.Update(ctx, p)
			}
			if err := s.activityRepo.RemoveProduct(ctx, activityID, ep.ProductID); err != nil {
				return err
			}
		}
	}

	// 添加或更新商品库存
	for _, id := range productIDs {
		stock := productStocks[id]
		var oldAllocated int64
		if _, ok := existingMap[id]; ok {
			for _, ep := range existing {
				if ep.ProductID == id {
					oldAllocated = ep.SeckillStock
					break
				}
			}
		}
		if stock <= 0 {
			p, err := s.productRepo.GetByID(ctx, id)
			if err != nil {
				continue
			}
			stock = p.Stock + oldAllocated // 默认用可用库存+已划拨库存
		}
		// 重新划拨库存：先把旧的归还，再按新值扣减
		p, err := s.productRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		available := p.Stock + oldAllocated
		if stock > available {
			stock = available
		}
		p.Stock = available - stock
		if err := s.productRepo.Update(ctx, p); err != nil {
			continue
		}
		if err := s.activityRepo.AddProduct(ctx, activityID, id, stock); err != nil {
			return err
		}
	}

	return nil
}

// GetActivity 获取活动详情（包含商品列表信息）
func (s *SeckillActivityService) GetActivity(ctx context.Context, id int64) (*ActivityDetail, error) {
	activity, err := s.activityRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	products, err := s.activityRepo.GetProductsByActivity(ctx, id)
	if err != nil {
		return nil, err
	}

	list := make([]*ProductInActivity, 0, len(products))
	for _, ap := range products {
		p, err := s.productRepo.GetByID(ctx, ap.ProductID)
		if err != nil {
			continue
		}
		list = append(list, &ProductInActivity{
			ProductID:    ap.ProductID,
			ProductName:  p.Name,
			ProductPrice: p.Price,
			SeckillStock: ap.SeckillStock,
			SeckillPrice: int64(float64(p.Price) * activity.Discount),
		})
	}

	return &ActivityDetail{
		Activity: activity,
		Products: list,
	}, nil
}

// ListActivities 列出所有活动
func (s *SeckillActivityService) ListActivities(ctx context.Context) ([]*seckill_activity.SeckillActivity, error) {
	return s.activityRepo.ListAll(ctx)
}

// GetActivityByProduct 根据商品ID获取该商品所属的活动信息
// 若有多个活动，优先返回当前进行中的活动，其次返回最近的一个活动
func (s *SeckillActivityService) GetActivityByProduct(ctx context.Context, productID int64) (*seckill_activity.SeckillActivity, error) {
	activities, err := s.activityRepo.GetActivitiesByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	if len(activities) == 0 {
		return nil, nil
	}

	now := time.Now()
	for _, act := range activities {
		if act.Status == 1 && now.After(act.StartTime) && now.Before(act.EndTime) {
			return act, nil
		}
	}
	// 没有进行中的活动时，返回第一个记录（通常是最近的一个）
	return activities[0], nil
}

// CheckAndUpdateExpiredActivities 检查并更新已过期的活动（恢复商品状态）
func (s *SeckillActivityService) CheckAndUpdateExpiredActivities(ctx context.Context) error {
	activities, err := s.activityRepo.ListAll(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, activity := range activities {
		// 结束时间已过且当前状态不是“已结束”时，更新状态并恢复商品
		if (now.After(activity.EndTime) || now.Equal(activity.EndTime)) && activity.Status != 2 {
			oldStatus := activity.Status
			activity.Status = 2
			if err := s.activityRepo.Update(ctx, activity); err != nil {
				// 记录错误但不影响其他活动
				_ = oldStatus
				continue
			}

			products, err := s.activityRepo.GetProductsByActivity(ctx, activity.ID)
			if err != nil {
				continue
			}
			for _, ap := range products {
				p, err := s.productRepo.GetByID(ctx, ap.ProductID)
				if err != nil {
					continue
				}
				if p.Status == 2 {
					p.Status = 1
					p.SeckillStock = 0
					if err := s.productRepo.Update(ctx, p); err != nil {
						continue
					}
				}
			}
		}
	}
	return nil
}

// CheckAndActivateStartedActivities 自动启动已到开始时间但尚未标记为进行中的活动
// 当当前时间位于 [StartTime, EndTime) 且活动状态不是 1 时，会调用 StartActivity
// 来同步活动状态、商品状态以及 Redis 中的秒杀库存。
func (s *SeckillActivityService) CheckAndActivateStartedActivities(ctx context.Context, seckillSvc *SeckillService) error {
	activities, err := s.activityRepo.ListAll(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, activity := range activities {
		if now.After(activity.StartTime) && now.Before(activity.EndTime) && activity.Status != 1 {
			if err := s.StartActivity(ctx, activity.ID, seckillSvc); err != nil {
				// 单个活动失败不影响其他活动
				continue
			}
		}
	}
	return nil
}

// DeleteActivity 删除活动
func (s *SeckillActivityService) DeleteActivity(ctx context.Context, id int64) error {
	// 归还所有已划拨的秒杀库存
	products, _ := s.activityRepo.GetProductsByActivity(ctx, id)
	for _, ap := range products {
		if p, err := s.productRepo.GetByID(ctx, ap.ProductID); err == nil {
			p.Stock += ap.SeckillStock
			_ = s.productRepo.Update(ctx, p)
		}
	}
	return s.activityRepo.Delete(ctx, id)
}

// StartActivity 启动活动（更新商品状态并同步库存到Redis）
// 一般由后台“启动”按钮或 CheckAndActivateStartedActivities 调用
func (s *SeckillActivityService) StartActivity(ctx context.Context, id int64, seckillSvc *SeckillService) error {
	activity, err := s.activityRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	if now.Before(activity.StartTime) {
		activity.Status = 0 // 未开始
	} else if now.After(activity.EndTime) {
		activity.Status = 2 // 已结束
	} else {
		activity.Status = 1 // 进行中
	}

	if err := s.activityRepo.Update(ctx, activity); err != nil {
		return err
	}

	// 活动处于进行中时，同步商品状态与库存
	if activity.Status == 1 {
		products, err := s.activityRepo.GetProductsByActivity(ctx, id)
		if err != nil {
			return err
		}

		for _, ap := range products {
			p, err := s.productRepo.GetByID(ctx, ap.ProductID)
			if err != nil {
				continue
			}

			p.Status = 2
			p.StartTime = activity.StartTime
			p.EndTime = activity.EndTime
			p.SeckillStock = ap.SeckillStock
			if err := s.productRepo.Update(ctx, p); err != nil {
				continue
			}

			// 同步库存到 Redis
			if err := seckillSvc.InitProductStock(ctx, p); err != nil {
				continue
			}
		}
	}

	return nil
}

// ----- 请求和响应结构 -----

type CreateActivityRequest struct {
	Name          string
	Description   string
	StartTime     time.Time
	EndTime       time.Time
	Discount      float64
	LimitPerUser  int64
	ProductIDs    []int64
	ProductStocks map[int64]int64 // 商品ID -> 秒杀库存
}

type UpdateActivityRequest struct {
	Name         string
	Description  string
	StartTime    time.Time
	EndTime      time.Time
	Discount     float64
	LimitPerUser int64
}

// ActivityDetail 后台使用的活动详情结构

type ActivityDetail struct {
	Activity *seckill_activity.SeckillActivity
	Products []*ProductInActivity
}

// ProductInActivity 活动中包含的商品信息

type ProductInActivity struct {
	ProductID    int64
	ProductName  string
	ProductPrice int64
	SeckillStock int64
	SeckillPrice int64 // 秒杀价 = 原价 * 折扣
}
