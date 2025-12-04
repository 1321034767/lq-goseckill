const state = {
  editingId: null,
  products: [],
};

const statusMap = ["下线", "正常", "秒杀中"];

const toastEl = document.getElementById("toast");
const productTableBody = document.getElementById("product-table");
const orderTableBody = document.getElementById("order-table");
const productForm = document.getElementById("product-form");

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
  let body = null;
  try {
    body = await response.json();
  } catch (err) {
    throw new Error("接口返回异常");
  }
  if (!response.ok || body.code !== 0) {
    throw new Error(body?.msg || `请求失败(${response.status})`);
  }
  return body.data || null;
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
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
    date.getDate()
  )}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
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
      (p) => `<tr>
        <td>${p.ID}</td>
        <td>${p.Name}</td>
        <td>¥${centsToYuan(p.Price).toFixed(2)}</td>
        <td>${p.Stock}</td>
        <td>${p.SeckillStock}</td>
        <td>${statusMap[p.Status] ?? "未知"}</td>
        <td>${formatDateTime(p.StartTime)}</td>
        <td>${formatDateTime(p.EndTime)}</td>
        <td>
          <button class="btn btn-sm btn-link" data-edit="${p.ID}">编辑</button>
        </td>
      </tr>`
    )
    .join("");
}

function renderOrders(list) {
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

productTableBody.addEventListener("click", (event) => {
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

(async function bootstrap() {
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
