const state = {
  editingId: null,
  editingProductId: null, // 行内编辑的商品ID
  products: [],
  orders: [],
  users: [],
  userOrders: [],
  selectedUserId: null,
  activities: [],
  editingActivityId: null,
};

const statusMap = ["下线", "正常", "秒杀中"];

const toastEl = document.getElementById("toast");
const productTableBody = document.getElementById("product-table");
const orderTableBody = document.getElementById("order-table");
const productForm = document.getElementById("product-form");
const userTableBody = document.getElementById("user-table");
const userOrderTableBody = document.getElementById("user-order-table");
const userOrdersTitle = document.getElementById("user-orders-title");
const sections = document.querySelectorAll(".admin-section");
const navLinks = document.querySelectorAll(".admin-sidebar-menu a[data-section]");

const nameInput = document.getElementById("name");
const priceInput = document.getElementById("price");
const descInput = document.getElementById("description");
const stockInput = document.getElementById("stock");
const seckillInput = document.getElementById("seckillStock");
const statusSelect = document.getElementById("status");
const startInput = document.getElementById("startTime");
const endInput = document.getElementById("endTime");
const formTitle = document.getElementById("form-title");
const cancelEditBtn = document.getElementById("cancel-edit");

// Chat DOM
const chatRecentListEl = document.getElementById("chat-recent-list");
const chatAllListEl = document.getElementById("chat-all-list");
const chatMessagesEl = document.getElementById("chat-message-list");
const chatInputEl = document.getElementById("chat-input");
const chatSendBtn = document.getElementById("chat-send-btn");
const chatActiveNameEl = document.getElementById("chat-active-name");
const chatActiveStatusEl = document.getElementById("chat-active-status");
const chatActiveAvatarEl = document.getElementById("chat-active-avatar");
const chatToggleBtn = document.getElementById("chat-toggle-btn");

function showToast(message, variant = "success") {
  toastEl.textContent = message;
  toastEl.className = `alert alert-${variant}`;
  toastEl.classList.remove("d-none");
  setTimeout(() => toastEl.classList.add("d-none"), 3000);
}

async function callApi(path, options = {}) {
  const opts = {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    ...options,
  };
  const response = await fetch(path, opts);
  // 非 2xx 直接抛错
  if (!response.ok) {
    let msg = `请求失败(${response.status})`;
    try {
      const errBody = await response.json();
      if (errBody && errBody.msg) msg = errBody.msg;
    } catch (_) {}
    throw new Error(msg);
  }

  // 2xx 情况尽量解析 JSON，解析失败也视为成功
  try {
    const body = await response.json();
    if (typeof body === "object" && body !== null && Object.prototype.hasOwnProperty.call(body, "code")) {
      if (body.code !== 0) {
        throw new Error(body?.msg || "请求失败");
      }
      return body.data || null;
    }
    return body;
  } catch (_) {
    // 返回非 JSON 内容但状态码 2xx，则视为成功
    return null;
  }
}

function centsToYuan(value) {
  return (value ?? 0) / 100;
}

function yuanToCents(value) {
  const num = Number(value);
  if (Number.isNaN(num)) {
    return 0;
  }
  return Math.round(num * 100);
}

function formatDateTime(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("zh-CN", { hour12: false });
}

function formatForInput(value) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (n) => `${n}`.padStart(2, "0");
  // 精确到秒
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
    date.getDate()
  )}T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

// 转义HTML，防止XSS
function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function renderProducts(list) {
  state.products = list;
  if (!list.length) {
    productTableBody.innerHTML =
      '<tr><td colspan="9" class="text-center text-muted">暂无数据</td></tr>';
    return;
  }
  productTableBody.innerHTML = list
    .map(
      (p) => {
        const isEditing = state.editingProductId === p.ID;
        if (isEditing) {
          // 编辑模式：显示输入框
          return `<tr data-product-id="${p.ID}" class="table-warning">
            <td>${p.ID}</td>
            <td><input type="text" class="form-control form-control-sm" value="${escapeHtml(p.Name)}" data-field="name"></td>
            <td>
              <div class="input-group input-group-sm">
                <span class="input-group-text">¥</span>
                <input type="number" class="form-control" value="${centsToYuan(p.Price).toFixed(2)}" step="0.01" min="0" data-field="price">
              </div>
            </td>
            <td><input type="number" class="form-control form-control-sm" value="${p.Stock}" min="0" data-field="stock"></td>
            <td><input type="number" class="form-control form-control-sm" value="${p.SeckillStock}" min="0" data-field="seckillStock"></td>
            <td>
              <select class="form-select form-select-sm" data-field="status">
                <option value="0" ${p.Status === 0 ? 'selected' : ''}>下线</option>
                <option value="1" ${p.Status === 1 ? 'selected' : ''}>正常</option>
                <option value="2" ${p.Status === 2 ? 'selected' : ''}>秒杀中</option>
              </select>
            </td>
            <td>${formatDateTime(p.StartTime)}</td>
            <td>${formatDateTime(p.EndTime)}</td>
            <td>
              <button class="btn btn-sm btn-success" data-save="${p.ID}">保存</button>
              <button class="btn btn-sm btn-secondary" data-cancel-edit="${p.ID}">取消</button>
            </td>
          </tr>`;
        } else {
          // 显示模式：显示文本和编辑按钮
          return `<tr data-product-id="${p.ID}">
            <td>${p.ID}</td>
            <td>${escapeHtml(p.Name)}</td>
            <td>¥${centsToYuan(p.Price).toFixed(2)}</td>
            <td>${p.Stock}</td>
            <td>${p.SeckillStock}</td>
            <td>${statusMap[p.Status] ?? "未知"}</td>
            <td>${formatDateTime(p.StartTime)}</td>
            <td>${formatDateTime(p.EndTime)}</td>
            <td>
              <button class="btn btn-sm btn-link" data-edit-inline="${p.ID}">编辑</button>
            </td>
          </tr>`;
        }
      }
    )
    .join("");
  
  // 绑定行内编辑事件
  bindInlineEditEvents();
}

function renderOrders(list) {
  state.orders = list;
  if (!list.length) {
    orderTableBody.innerHTML =
      '<tr><td colspan="6" class="text-center text-muted">暂无订单</td></tr>';
    return;
  }
  orderTableBody.innerHTML = list
    .map(
      (o) => `<tr>
        <td>${o.ID}</td>
        <td>${o.UserID}</td>
        <td>${o.ProductID}</td>
        <td>¥${centsToYuan(o.Price).toFixed(2)}</td>
        <td>${o.Status}</td>
        <td>${formatDateTime(o.CreatedAt)}</td>
      </tr>`
    )
    .join("");
}

function renderUsers(list) {
  state.users = list;
  if (!list.length) {
    userTableBody.innerHTML =
      '<tr><td colspan="5" class="text-center text-muted">暂无用户</td></tr>';
    return;
  }
  userTableBody.innerHTML = list
    .map(
      (u) => `<tr data-user-id="${u.user_id}">
        <td>${u.user_id}</td>
        <td>${u.username}</td>
        <td>¥${centsToYuan(u.balance).toFixed(2)}</td>
        <td>¥${centsToYuan(u.frozen).toFixed(2)}</td>
        <td class="text-center">
          <button class="btn btn-sm btn-link" data-user="${u.user_id}" data-action="orders">查看订单</button>
          <button class="btn btn-sm btn-link text-success" data-user="${u.user_id}" data-action="recharge">充值</button>
        </td>
      </tr>`
    )
    .join("");
}

function renderUserOrders(list) {
  state.userOrders = list;
  if (!list.length) {
    userOrderTableBody.innerHTML =
      '<tr><td colspan="5" class="text-center text-muted">该用户暂无订单</td></tr>';
    return;
  }
  userOrderTableBody.innerHTML = list
    .map(
      (o) => `<tr>
        <td>${o.ID}</td>
        <td>${o.ProductID}</td>
        <td>¥${centsToYuan(o.Price).toFixed(2)}</td>
        <td>${o.Status}</td>
        <td>${formatDateTime(o.CreatedAt)}</td>
      </tr>`
    )
    .join("");
}

function switchSection(sectionId) {
  sections.forEach((sec) => {
    if (sec.id === sectionId) {
      sec.classList.remove("d-none");
    } else {
      sec.classList.add("d-none");
    }
  });
  navLinks.forEach((link) => {
    if (link.dataset.section === sectionId) {
      link.classList.add("active");
    } else {
      link.classList.remove("active");
    }
  });
  if (sectionId === "product-section") {
    loadProducts();
  } else if (sectionId === "order-section") {
    loadOrders();
  } else if (sectionId === "user-section") {
    loadUsers();
  } else if (sectionId === "seckill-activity-section") {
    loadActivities();
    loadProducts(); // 需要加载商品列表用于选择
  }
}

navLinks.forEach((link) => {
  link.addEventListener("click", (e) => {
    e.preventDefault();
    const target = link.dataset.section;
    if (target) switchSection(target);
  });
});

function resetForm() {
  state.editingId = null;
  productForm.reset();
  formTitle.textContent = "新建商品";
  cancelEditBtn.classList.add("d-none");
  document
    .querySelectorAll("#product-form input[required]")
    .forEach((input) => input.removeAttribute("disabled"));
}

function fillForm(product) {
  state.editingId = product.ID;
  formTitle.textContent = `编辑商品 #${product.ID}`;
  nameInput.value = product.Name;
  priceInput.value = centsToYuan(product.Price).toFixed(2);
  descInput.value = product.Description || "";
  stockInput.value = product.Stock;
  seckillInput.value = product.SeckillStock;
  statusSelect.value = product.Status;
  startInput.value = formatForInput(product.StartTime);
  endInput.value = formatForInput(product.EndTime);
  cancelEditBtn.classList.remove("d-none");
}

async function loadProducts() {
  productTableBody.innerHTML =
    '<tr><td colspan="9" class="text-center text-muted">加载中...</td></tr>';
  try {
    const data = await callApi("/api/products");
    renderProducts(Array.isArray(data) ? data : []);
  } catch (err) {
    showToast(err.message, "danger");
    productTableBody.innerHTML =
      '<tr><td colspan="9" class="text-center text-danger">加载失败</td></tr>';
  }
}

// 定时刷新商品列表（每15秒刷新一次，确保商品状态实时更新）
if (productTableBody) {
  setInterval(function() {
    // 只在商品列表可见时才刷新
    const productSection = document.getElementById("product-section");
    if (productSection && productSection.offsetParent !== null) {
      loadProducts();
    }
  }, 15000); // 15秒刷新一次
}

async function loadOrders() {
  orderTableBody.innerHTML =
    '<tr><td colspan="6" class="text-center text-muted">加载中...</td></tr>';
  try {
    const data = await callApi("/api/orders?limit=20");
    renderOrders(Array.isArray(data) ? data : []);
  } catch (err) {
    showToast(err.message, "danger");
    orderTableBody.innerHTML =
      '<tr><td colspan="6" class="text-center text-danger">加载失败</td></tr>';
  }
}

async function loadUsers() {
  if (!userTableBody) return;
  userTableBody.innerHTML =
    '<tr><td colspan="5" class="text-center text-muted">加载中...</td></tr>';
  try {
    const data = await callApi("/api/users");
    renderUsers(Array.isArray(data) ? data : []);
  } catch (err) {
    showToast(err.message, "danger");
    userTableBody.innerHTML =
      '<tr><td colspan="5" class="text-center text-danger">加载失败</td></tr>';
  }
}

async function loadUserOrders(userID) {
  if (!userOrderTableBody) return;
  state.selectedUserId = userID;
  if (userOrdersTitle) {
    userOrdersTitle.textContent = `用户 ${userID} 的订单`;
  }
  userOrderTableBody.innerHTML =
    '<tr><td colspan="5" class="text-center text-muted">加载中...</td></tr>';
  try {
    const data = await callApi(`/api/users/${userID}/orders`);
    renderUserOrders(Array.isArray(data) ? data : []);
  } catch (err) {
    showToast(err.message, "danger");
    userOrderTableBody.innerHTML =
      '<tr><td colspan="5" class="text-center text-danger">加载失败</td></tr>';
  }
}

// 给用户充值
async function rechargeUser(userID) {
  const input = prompt(`请输入给用户 ${userID} 充值的金额（元）：`, "100.00");
  if (input === null) return;
  const amountYuan = parseFloat(input);
  if (Number.isNaN(amountYuan) || amountYuan <= 0) {
    showToast("充值金额必须大于 0", "danger");
    return;
  }
  const amountCents = yuanToCents(amountYuan);
  try {
    await callApi(`/api/users/${userID}/recharge`, {
      method: "POST",
      body: JSON.stringify({ amount: amountCents }),
    });
    showToast(`用户 ${userID} 充值成功`, "success");
    await loadUsers();
  } catch (err) {
    showToast(err.message, "danger");
  }
}

productForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = {
    name: nameInput.value.trim(),
    description: descInput.value.trim(),
    price: yuanToCents(priceInput.value),
    stock: Number(stockInput.value),
    seckill_stock: Number(seckillInput.value),
    status: Number(statusSelect.value),
    start_time: startInput.value
      ? new Date(startInput.value).toISOString()
      : "",
    end_time: endInput.value ? new Date(endInput.value).toISOString() : "",
  };

  const method = state.editingId ? "PUT" : "POST";
  const url = state.editingId
    ? `/api/products/${state.editingId}`
    : "/api/products";

  try {
    await callApi(url, {
      method,
      body: JSON.stringify(payload),
    });
    showToast("保存成功");
    await loadProducts();
    resetForm();
  } catch (err) {
    showToast(err.message, "danger");
  }
});

document
  .getElementById("refresh-products")
  .addEventListener("click", loadProducts);

document.getElementById("refresh-orders").addEventListener("click", loadOrders);

document.getElementById("reset-form").addEventListener("click", resetForm);
cancelEditBtn.addEventListener("click", resetForm);

// 绑定行内编辑事件
function bindInlineEditEvents() {
  // 行内编辑按钮点击事件
  productTableBody.querySelectorAll("button[data-edit-inline]").forEach(btn => {
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      const id = Number(btn.dataset.editInline);
      startInlineEdit(id);
    });
  });

  // 保存按钮点击事件
  productTableBody.querySelectorAll("button[data-save]").forEach(btn => {
    btn.addEventListener("click", async (e) => {
      e.stopPropagation();
      const id = Number(btn.dataset.save);
      await saveInlineEdit(id);
    });
  });

  // 取消编辑按钮点击事件
  productTableBody.querySelectorAll("button[data-cancel-edit]").forEach(btn => {
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      cancelInlineEdit();
    });
  });
}

// 开始行内编辑
function startInlineEdit(productId) {
  // 如果已经有商品在编辑，先取消
  if (state.editingProductId !== null && state.editingProductId !== productId) {
    cancelInlineEdit();
  }
  state.editingProductId = productId;
  renderProducts(state.products);
}

// 取消行内编辑
function cancelInlineEdit() {
  state.editingProductId = null;
  renderProducts(state.products);
}

// 保存行内编辑
async function saveInlineEdit(productId) {
  const row = productTableBody.querySelector(`tr[data-product-id="${productId}"]`);
  if (!row) {
    showToast("未找到商品行", "danger");
    return;
  }

  // 收集表单数据
  const nameInput = row.querySelector('input[data-field="name"]');
  const priceInput = row.querySelector('input[data-field="price"]');
  const stockInput = row.querySelector('input[data-field="stock"]');
  const seckillStockInput = row.querySelector('input[data-field="seckillStock"]');
  const statusSelect = row.querySelector('select[data-field="status"]');

  if (!nameInput || !priceInput || !stockInput || !seckillStockInput || !statusSelect) {
    showToast("表单字段不完整", "danger");
    return;
  }

  const name = nameInput.value.trim();
  const price = parseFloat(priceInput.value);
  const stock = parseInt(stockInput.value, 10);
  const seckillStock = parseInt(seckillStockInput.value, 10);
  const status = parseInt(statusSelect.value, 10);

  // 验证数据
  if (!name) {
    showToast("商品名称不能为空", "danger");
    return;
  }
  if (isNaN(price) || price < 0) {
    showToast("价格无效", "danger");
    return;
  }
  if (isNaN(stock) || stock < 0) {
    showToast("库存无效", "danger");
    return;
  }
  if (isNaN(seckillStock) || seckillStock < 0) {
    showToast("秒杀库存无效", "danger");
    return;
  }

  // 获取原始商品数据
  const product = state.products.find(p => p.ID === productId);
  if (!product) {
    showToast("未找到商品", "danger");
    return;
  }

  // 格式化时间
  const formatTimeForAPI = (timeStr) => {
    if (!timeStr) return "";
    const date = new Date(timeStr);
    if (isNaN(date.getTime())) return "";
    return date.toISOString().slice(0, 19).replace("T", " ");
  };

  // 构建更新请求
  const updateData = {
    name: name,
    price: Math.round(price * 100), // 转换为分
    stock: stock,
    seckill_stock: seckillStock,
    status: status,
    description: product.Description || "",
    category: product.Category || "",
    start_time: formatTimeForAPI(product.StartTime),
    end_time: formatTimeForAPI(product.EndTime),
  };

  try {
    // 禁用保存按钮，显示加载状态
    const saveBtn = row.querySelector('button[data-save]');
    if (saveBtn) {
      saveBtn.disabled = true;
      saveBtn.textContent = "保存中...";
    }

    const result = await callApi(`/api/products/${productId}`, {
      method: "PUT",
      body: JSON.stringify(updateData),
    });
    // callApi 已在非 2xx 时抛异常，这里直接视为成功
    const updatedProduct = result || {};

    showToast("商品更新成功", "success");
    
    // 更新本地状态
    const index = state.products.findIndex(p => p.ID === productId);
    if (index !== -1) {
      state.products[index] = updatedProduct;
    }

    // 取消编辑状态并重新渲染
    state.editingProductId = null;
    await loadProducts();
  } catch (err) {
    showToast(err.message || "更新失败", "danger");
    // 恢复保存按钮
    const saveBtn = row.querySelector('button[data-save]');
    if (saveBtn) {
      saveBtn.disabled = false;
      saveBtn.textContent = "保存";
    }
  }
}

productTableBody.addEventListener("click", (event) => {
  // 行内编辑按钮（新）
  const inlineEditBtn = event.target.closest("button[data-edit-inline]");
  if (inlineEditBtn) {
    const id = Number(inlineEditBtn.dataset.editInline);
    startInlineEdit(id);
    return;
  }

  // 原有的表单编辑按钮（保留兼容）
  const btn = event.target.closest("button[data-edit]");
  if (!btn) return;
  const id = Number(btn.dataset.edit);
  const product = state.products.find((item) => item.ID === id);
  if (!product) {
    showToast("未找到该商品", "danger");
    return;
  }
  fillForm(product);
});

const refreshUsersBtn = document.getElementById("refresh-users");
if (refreshUsersBtn) {
  refreshUsersBtn.addEventListener("click", loadUsers);
}
if (userTableBody) {
  userTableBody.addEventListener("click", (event) => {
    const btn = event.target.closest("button[data-user]");
    if (!btn) return;
    const uid = Number(btn.dataset.user);
    if (Number.isNaN(uid)) return;
    const action = btn.dataset.action || "orders";
    if (action === "orders") {
      loadUserOrders(uid);
    } else if (action === "recharge") {
      rechargeUser(uid);
    }
  });
}

(async function bootstrap() {
  switchSection("product-section");
  await Promise.all([loadProducts(), loadOrders()]);
  initChat();
})();

// ---------- Chat 逻辑：通过后端接口实现真实存储 ----------

const chatState = {
  activeId: "claire",
  contacts: [],
  messages: {}, // { contactId: [{id, from, content}] }
  lastIds: {}, // { contactId: lastMessageID }
};

async function loadChatContacts() {
  try {
    const data = await callApi("/api/chat/contacts");
    chatState.contacts = Array.isArray(data) ? data : [];
  } catch (err) {
    showToast(`加载联系人失败: ${err.message}`, "danger");
  }
}

async function loadChatMessages(contactId, { initial } = { initial: false }) {
  const afterID = initial ? 0 : chatState.lastIds[contactId] || 0;
  const qs = afterID ? `?after_id=${afterID}&limit=50` : "?limit=50";
  try {
    const data = await callApi(`/api/chat/messages/${contactId}${qs}`);
    const list = Array.isArray(data) ? data : [];
    if (!chatState.messages[contactId] || initial) {
      chatState.messages[contactId] = [];
    }
    if (list.length) {
      chatState.messages[contactId].push(
        ...list.map((m) => ({
          id: m.ID,
          from: m.From,
          text: m.Content,
        }))
      );
      chatState.lastIds[contactId] = list[list.length - 1].ID;
    }
  } catch (err) {
    if (initial) {
      showToast(`加载聊天记录失败: ${err.message}`, "danger");
    }
  }
}

function avatarChar(name) {
  if (!name) return "?";
  return name.trim().charAt(0).toUpperCase();
}

function statusClass(status) {
  if (status === "online") return "status-online";
  if (status === "away") return "status-away";
  return "status-offline";
}

function renderChatContacts() {
  if (!chatRecentListEl || !chatAllListEl) return;
  const recents = chatState.contacts.slice(0, 3);
  const others = chatState.contacts.slice(3);

  chatRecentListEl.innerHTML = recents.map(renderContactItem).join("");
  chatAllListEl.innerHTML = others.map(renderContactItem).join("");

  function renderContactItem(c) {
    const active = c.id === chatState.activeId ? " active" : "";
    const preview =
      c.lastMessage && c.lastMessage.length > 22
        ? `${c.lastMessage.slice(0, 22)}...`
        : c.lastMessage || "";
    return `<div class="admin-chat-contact${active}" data-chat-id="${c.id}">
      <div class="avatar-circle avatar-sm">${avatarChar(c.name)}</div>
      <div class="flex-grow-1">
        <div class="d-flex justify-content-between align-items-center">
          <span class="admin-chat-contact-name">${c.name}</span>
          <span class="admin-chat-status-dot ${statusClass(
            c.status
          )}"></span>
        </div>
        ${
          preview
            ? `<div class="admin-chat-contact-message">${preview}</div>`
            : ""
        }
      </div>
    </div>`;
  }
}

function renderChatMessages() {
  if (!chatMessagesEl) return;
  const list = chatState.messages[chatState.activeId] || [];
  if (!list.length) {
    chatMessagesEl.innerHTML =
      '<div class="admin-chat-empty">还没有消息，先打个招呼吧～</div>';
    return;
  }
  chatMessagesEl.innerHTML = list
    .map(
      (m) => `<div class="admin-chat-message-row ${
        m.from === "self" ? "self" : "friend"
      }">
      <div class="admin-chat-message">${m.text}</div>
    </div>`
    )
    .join("");
  chatMessagesEl.scrollTop = chatMessagesEl.scrollHeight;
}

function renderChatHeader() {
  if (!chatActiveNameEl || !chatActiveStatusEl || !chatActiveAvatarEl) return;
  const contact =
    chatState.contacts.find((c) => c.id === chatState.activeId) ||
    chatState.contacts[0];
  if (!contact) return;
  chatActiveNameEl.textContent = contact.name;
  chatActiveAvatarEl.textContent = avatarChar(contact.name);
  const statusText =
    contact.status === "online"
      ? "Online"
      : contact.status === "away"
      ? "Active 1h ago"
      : "Offline";
  chatActiveStatusEl.textContent = statusText;
}

function switchChat(id) {
  chatState.activeId = id;
  renderChatContacts();
  renderChatHeader();
  loadChatMessages(id, { initial: true }).then(() => {
    renderChatMessages();
  });
}

async function sendChatMessage() {
  if (!chatInputEl) return;
  const text = chatInputEl.value.trim();
  if (!text) return;
  const id = chatState.activeId;
  chatInputEl.value = "";

  try {
    const data = await callApi(`/api/chat/messages/${id}`, {
      method: "POST",
      body: JSON.stringify({ content: text }),
    });
    // 后端返回新消息，写入本地状态
    if (!chatState.messages[id]) {
      chatState.messages[id] = [];
    }
    const msg = {
      id: data.ID,
      from: data.From,
      text: data.Content,
    };
    chatState.messages[id].push(msg);
    chatState.lastIds[id] = msg.id;
    renderChatMessages();
  } catch (err) {
    showToast(`发送消息失败: ${err.message}`, "danger");
  }
}

function initChat() {
  if (!chatMessagesEl) return;
  // 先加载联系人，再加载默认会话的历史消息
  loadChatContacts()
    .then(() => {
      if (!chatState.contacts.length) {
        return;
      }
      // 使用第一个联系人作为默认会话
      chatState.activeId = chatState.contacts[0].id;
      renderChatContacts();
      renderChatHeader();
      return loadChatMessages(chatState.activeId, { initial: true });
    })
    .then(() => {
      renderChatMessages();
    });

  [chatRecentListEl, chatAllListEl].forEach((el) => {
    if (!el) return;
    el.addEventListener("click", (event) => {
      const item = event.target.closest("[data-chat-id]");
      if (!item) return;
      switchChat(item.getAttribute("data-chat-id"));
    });
  });

  if (chatSendBtn) {
    chatSendBtn.addEventListener("click", sendChatMessage);
  }
  if (chatInputEl) {
    chatInputEl.addEventListener("keydown", (event) => {
      if (event.key === "Enter") {
        event.preventDefault();
        sendChatMessage();
      }
    });
  }

  if (chatToggleBtn) {
    chatToggleBtn.addEventListener("click", () => {
      document.body.classList.toggle("show-chat");
    });
  }
}

// ---------- 秒杀活动管理逻辑 ----------

const activityStatusMap = ["未开始", "进行中", "已结束", "已取消"];

const activityTableBody = document.getElementById("activity-table");
const activityFormCard = document.getElementById("activity-form-card");
const activityForm = document.getElementById("activity-form");
const activityFormTitle = document.getElementById("activity-form-title");
const productSelectList = document.getElementById("product-select-list");
const activityDetailCard = document.getElementById("activity-detail-card");
const activityDetailContent = document.getElementById("activity-detail-content");
const activityDetailTitle = document.getElementById("activity-detail-title");
const activityDetailSubtitle = document.getElementById("activity-detail-subtitle");

function renderActivities(list) {
  state.activities = list;
  if (!list.length) {
    activityTableBody.innerHTML =
      '<tr><td colspan="7" class="text-center text-muted">暂无活动</td></tr>';
    return;
  }
  activityTableBody.innerHTML = list
    .map(
      (a) => {
        const now = new Date();
        const startTime = new Date(a.StartTime);
        const endTime = new Date(a.EndTime);
        let status = a.Status;
        // 优先使用后端返回的状态，但如果时间已过，强制更新为已结束
        if (now > endTime) {
          status = 2; // 已结束
        } else if (status === 0 && now >= startTime && now <= endTime) {
          status = 1; // 自动判断为进行中
        }
        return `<tr>
          <td>${a.ID}</td>
          <td>${a.Name}</td>
          <td>${(a.Discount * 10).toFixed(1)}折</td>
          <td>${formatDateTime(a.StartTime)}</td>
          <td>${formatDateTime(a.EndTime)}</td>
          <td>每人限购${a.LimitPerUser || 1}件</td>
          <td>${activityStatusMap[status] || "未知"}</td>
          <td class="text-center">
            <button class="btn btn-sm btn-link" data-view-activity="${a.ID}">查看</button>
            <button class="btn btn-sm btn-link text-primary" data-start-activity="${a.ID}">启动</button>
            <button class="btn btn-sm btn-link text-danger" data-delete-activity="${a.ID}">删除</button>
          </td>
        </tr>`;
      }
    )
    .join("");
}

function renderProductSelectList() {
  if (!state.products.length) {
    productSelectList.innerHTML = '<div class="col-12 text-muted">暂无商品</div>';
    return;
  }
  productSelectList.innerHTML = state.products
    .map(
      (p) => `<div class="col-md-6 col-lg-4">
        <div class="form-check border rounded p-2">
          <input class="form-check-input" type="checkbox" value="${p.ID}" id="product-${p.ID}" data-product-id="${p.ID}">
          <label class="form-check-label w-100" for="product-${p.ID}">
            <div class="fw-semibold">${p.Name}</div>
            <small class="text-muted">¥${centsToYuan(p.Price).toFixed(2)} | 库存: ${p.Stock}</small>
            <div class="mt-1">
              <small>秒杀库存:</small>
              <input type="number" class="form-control form-control-sm d-inline-block w-auto ms-1" 
                     data-stock-input="${p.ID}" 
                     value="${p.Stock}" 
                     min="0" 
                     max="${p.Stock}"
                     style="width: 80px;">
            </div>
          </label>
        </div>
      </div>`
    )
    .join("");
}

async function loadActivities() {
  if (!activityTableBody) return;
  activityTableBody.innerHTML =
    '<tr><td colspan="7" class="text-center text-muted">加载中...</td></tr>';
  try {
    const data = await callApi("/api/seckill-activities");
    renderActivities(Array.isArray(data) ? data : []);
  } catch (err) {
    showToast(err.message, "danger");
    activityTableBody.innerHTML =
      '<tr><td colspan="7" class="text-center text-danger">加载失败</td></tr>';
  }
}

function resetActivityForm() {
  state.editingActivityId = null;
  activityForm.reset();
  activityFormTitle.textContent = "新建秒杀活动";
  activityFormCard.style.display = "none";
  // 重置限购数量为默认值1
  const limitInput = document.getElementById("activity-limit-per-user");
  if (limitInput) {
    limitInput.value = "1";
  }
  // 重置商品选择
  if (productSelectList) {
    productSelectList.querySelectorAll("input[type='checkbox']").forEach((cb) => {
      cb.checked = false;
    });
    productSelectList.querySelectorAll("input[data-stock-input]").forEach((input) => {
      const productId = input.getAttribute("data-stock-input");
      const product = state.products.find((p) => p.ID === Number(productId));
      if (product) {
        input.value = product.Stock;
      }
    });
  }
}

function showActivityForm() {
  activityFormCard.style.display = "block";
  renderProductSelectList();
  // 设置默认时间为1小时后开始，2小时后结束
  const now = new Date();
  const startTime = new Date(now.getTime() + 60 * 60 * 1000);
  const endTime = new Date(now.getTime() + 2 * 60 * 60 * 1000);
  document.getElementById("activity-start-time").value = formatForInput(startTime);
  document.getElementById("activity-end-time").value = formatForInput(endTime);
  // 设置默认限购数量
  document.getElementById("activity-limit-per-user").value = "1";
}

async function showActivityDetail(activityId) {
  try {
    const data = await callApi(`/api/seckill-activities/${activityId}`);
    if (!data || !data.Activity) {
      showToast("活动不存在", "danger");
      return;
    }
    const activity = data.Activity;
    const products = data.Products || [];
    
    activityDetailTitle.textContent = activity.Name;
    activityDetailSubtitle.textContent = `${formatDateTime(activity.StartTime)} - ${formatDateTime(activity.EndTime)} | ${(activity.Discount * 10).toFixed(1)}折 | 每人限购${activity.LimitPerUser || 1}件`;
    
    if (!products.length) {
      activityDetailContent.innerHTML = '<p class="text-muted">该活动暂无商品</p>';
    } else {
      activityDetailContent.innerHTML = `
        <table class="table table-striped">
          <thead>
            <tr>
              <th>商品ID</th>
              <th>商品名称</th>
              <th>原价</th>
              <th>秒杀价</th>
              <th>秒杀库存</th>
            </tr>
          </thead>
          <tbody>
            ${products.map(p => `
              <tr>
                <td>${p.ProductID}</td>
                <td>${p.ProductName}</td>
                <td>¥${centsToYuan(p.ProductPrice).toFixed(2)}</td>
                <td class="text-danger fw-bold">¥${centsToYuan(p.SeckillPrice).toFixed(2)}</td>
                <td>${p.SeckillStock}</td>
              </tr>
            `).join("")}
          </tbody>
        </table>
      `;
    }
    
    activityDetailCard.style.display = "block";
    activityFormCard.style.display = "none";
  } catch (err) {
    showToast(err.message, "danger");
  }
}

if (activityForm) {
  activityForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    
    const name = document.getElementById("activity-name").value.trim();
    const description = document.getElementById("activity-description").value.trim();
    const startTime = document.getElementById("activity-start-time").value;
    const endTime = document.getElementById("activity-end-time").value;
    const discount = parseFloat(document.getElementById("activity-discount").value);
    const limitPerUser = parseInt(document.getElementById("activity-limit-per-user").value) || 1;
    
    if (!name) {
      showToast("请输入活动名称", "danger");
      return;
    }
    
    if (!startTime || !endTime) {
      showToast("请选择开始和结束时间", "danger");
      return;
    }
    
    if (new Date(startTime) >= new Date(endTime)) {
      showToast("结束时间必须晚于开始时间", "danger");
      return;
    }
    
    if (discount <= 0 || discount > 1) {
      showToast("折扣必须在0.1-1.0之间", "danger");
      return;
    }
    
    // 获取选中的商品
    const selectedProducts = Array.from(productSelectList.querySelectorAll("input[type='checkbox']:checked"))
      .map(cb => Number(cb.value));
    
    if (selectedProducts.length === 0) {
      showToast("请至少选择一个商品", "danger");
      return;
    }
    
    // 获取每个商品的秒杀库存
    const productStocks = {};
    selectedProducts.forEach(productId => {
      const stockInput = productSelectList.querySelector(`input[data-stock-input="${productId}"]`);
      if (stockInput) {
        productStocks[productId] = Number(stockInput.value) || 0;
      }
    });
    
    const payload = {
      name: name,
      description: description,
      start_time: new Date(startTime).toISOString(),
      end_time: new Date(endTime).toISOString(),
      discount: discount,
      limit_per_user: limitPerUser,
      product_ids: selectedProducts,
      product_stocks: productStocks,
    };
    
    try {
      await callApi("/api/seckill-activities", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      showToast("活动创建成功");
      await loadActivities();
      resetActivityForm();
    } catch (err) {
      showToast(err.message, "danger");
    }
  });
}

if (document.getElementById("new-activity-btn")) {
  document.getElementById("new-activity-btn").addEventListener("click", () => {
    resetActivityForm();
    showActivityForm();
  });
}

if (document.getElementById("refresh-activities")) {
  document.getElementById("refresh-activities").addEventListener("click", loadActivities);
  
  // 定时刷新活动列表（每20秒刷新一次，确保活动状态实时更新）
  setInterval(function() {
    const activitySection = document.getElementById("seckill-activity-section");
    if (activitySection && activitySection.offsetParent !== null) {
      loadActivities();
    }
  }, 20000); // 20秒刷新一次
}

if (document.getElementById("cancel-activity-edit")) {
  document.getElementById("cancel-activity-edit").addEventListener("click", resetActivityForm);
}

if (document.getElementById("close-activity-detail")) {
  document.getElementById("close-activity-detail").addEventListener("click", () => {
    activityDetailCard.style.display = "none";
  });
}

if (activityTableBody) {
  activityTableBody.addEventListener("click", async (event) => {
    const viewBtn = event.target.closest("button[data-view-activity]");
    if (viewBtn) {
      const id = Number(viewBtn.dataset.viewActivity);
      await showActivityDetail(id);
      return;
    }
    
    const startBtn = event.target.closest("button[data-start-activity]");
    if (startBtn) {
      const id = Number(startBtn.dataset.startActivity);
      if (!confirm("确定要启动这个秒杀活动吗？这将更新商品状态并同步库存到Redis。")) {
        return;
      }
      try {
        await callApi(`/api/seckill-activities/${id}/start`, { method: "POST" });
        showToast("活动启动成功");
        await loadActivities();
      } catch (err) {
        showToast(err.message, "danger");
      }
      return;
    }
    
    const deleteBtn = event.target.closest("button[data-delete-activity]");
    if (deleteBtn) {
      const id = Number(deleteBtn.dataset.deleteActivity);
      if (!confirm("确定要删除这个秒杀活动吗？")) {
        return;
      }
      try {
        await callApi(`/api/seckill-activities/${id}`, { method: "DELETE" });
        showToast("活动删除成功");
        await loadActivities();
      } catch (err) {
        showToast(err.message, "danger");
      }
      return;
    }
  });
}
