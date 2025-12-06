package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/example/goseckill/internal/datamodels/seckill_activity"
)

type seckillActivityRepo struct {
	db *gorm.DB
}

// NewSeckillActivityRepository 创建秒杀活动仓储
func NewSeckillActivityRepository(db *gorm.DB) seckill_activity.Repository {
	return &seckillActivityRepo{db: db}
}

func (r *seckillActivityRepo) Create(ctx context.Context, activity *seckill_activity.SeckillActivity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

func (r *seckillActivityRepo) GetByID(ctx context.Context, id int64) (*seckill_activity.SeckillActivity, error) {
	var activity seckill_activity.SeckillActivity
	if err := r.db.WithContext(ctx).First(&activity, id).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *seckillActivityRepo) ListAll(ctx context.Context) ([]*seckill_activity.SeckillActivity, error) {
	var list []*seckill_activity.SeckillActivity
	if err := r.db.WithContext(ctx).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *seckillActivityRepo) Update(ctx context.Context, activity *seckill_activity.SeckillActivity) error {
	return r.db.WithContext(ctx).Save(activity).Error
}

func (r *seckillActivityRepo) Delete(ctx context.Context, id int64) error {
	// 先删除关联的商品
	if err := r.db.WithContext(ctx).Where("activity_id = ?", id).Delete(&seckill_activity.SeckillActivityProduct{}).Error; err != nil {
		return err
	}
	// 再删除活动
	return r.db.WithContext(ctx).Delete(&seckill_activity.SeckillActivity{}, id).Error
}

func (r *seckillActivityRepo) AddProduct(ctx context.Context, activityID, productID, seckillStock int64) error {
	// 检查是否已存在
	var existing seckill_activity.SeckillActivityProduct
	err := r.db.WithContext(ctx).Where("activity_id = ? AND product_id = ?", activityID, productID).First(&existing).Error
	if err == nil {
		// 已存在，更新库存
		existing.SeckillStock = seckillStock
		return r.db.WithContext(ctx).Save(&existing).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	// 不存在，创建新记录
	ap := &seckill_activity.SeckillActivityProduct{
		ActivityID:   activityID,
		ProductID:    productID,
		SeckillStock: seckillStock,
	}
	return r.db.WithContext(ctx).Create(ap).Error
}

func (r *seckillActivityRepo) RemoveProduct(ctx context.Context, activityID, productID int64) error {
	return r.db.WithContext(ctx).
		Where("activity_id = ? AND product_id = ?", activityID, productID).
		Delete(&seckill_activity.SeckillActivityProduct{}).Error
}

func (r *seckillActivityRepo) GetProductsByActivity(ctx context.Context, activityID int64) ([]*seckill_activity.SeckillActivityProduct, error) {
	var list []*seckill_activity.SeckillActivityProduct
	if err := r.db.WithContext(ctx).
		Where("activity_id = ?", activityID).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *seckillActivityRepo) GetActivitiesByProduct(ctx context.Context, productID int64) ([]*seckill_activity.SeckillActivity, error) {
	var activities []*seckill_activity.SeckillActivity
	if err := r.db.WithContext(ctx).
		Table("seckill_activities").
		Joins("INNER JOIN seckill_activity_products ON seckill_activities.id = seckill_activity_products.activity_id").
		Where("seckill_activity_products.product_id = ?", productID).
		Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}
