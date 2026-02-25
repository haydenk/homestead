"use strict";

// ── State ────────────────────────────────────────────────────────────────────
let config   = {};
let statuses = {};

const state = { query: "" };

// ── Selectors ────────────────────────────────────────────────────────────────
const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => [...document.querySelectorAll(sel)];

const el = {
  title:      () => $("#site-title"),
  subtitle:   () => $("#site-subtitle"),
  logo:       () => $("#site-logo"),
  footer:     () => $("#footer-text"),
  sections:   () => $("#sections"),
  loading:    () => $("#loading-state"),
  empty:      () => $("#empty-state"),
  emptyQuery: () => $("#empty-query"),
  search:     () => $("#search"),
  toast:      () => $("#toast"),
  btnRefresh: () => $("#btn-refresh"),
};

// ── Init ─────────────────────────────────────────────────────────────────────
document.addEventListener("DOMContentLoaded", async () => {
  bindEvents();

  try {
    await loadConfig();
    renderDashboard();
    await loadStatus();
    startPolling();
  } catch (err) {
    console.error("Init error:", err);
    showToast("⚠ Could not load dashboard config.", 5000);
  } finally {
    el.loading().hidden = true;
  }
});

// ── Data fetching ─────────────────────────────────────────────────────────────
async function loadConfig() {
  const res = await fetch("/api/config");
  if (!res.ok) throw new Error(`config fetch: ${res.status}`);
  config = await res.json();
}

async function loadStatus() {
  try {
    const res = await fetch("/api/status");
    if (!res.ok) return;
    statuses = await res.json();
    updateStatusDots();
  } catch (e) {
    console.warn("status fetch:", e);
  }
}

// ── Render ────────────────────────────────────────────────────────────────────
function renderDashboard() {
  document.title = config.title || "Homestead";
  el.title().textContent    = config.title    || "Homestead";
  el.subtitle().textContent = config.subtitle || "";
  el.logo().textContent     = config.logo     || "🏠";
  el.footer().textContent   = config.footer   || "";

  const cols = Math.min(Math.max(config.columns || 4, 1), 6);
  document.documentElement.style.setProperty("--columns", cols);

  el.sections().innerHTML = (config.sections || []).map(renderSection).join("");
}

function renderSection(section) {
  const items = section.items || [];
  return `
    <section class="section" data-section="${esc(section.name)}">
      <div class="section-header">
        ${section.icon ? `<span class="section-icon" aria-hidden="true">${esc(section.icon)}</span>` : ""}
        <span class="section-name">${esc(section.name)}</span>
        <span class="section-count" aria-label="${items.length} items">${items.length}</span>
      </div>
      <div class="cards-grid">
        ${items.map(renderCard).join("")}
      </div>
    </section>`;
}

function renderCard(item) {
  const accentStyle = item.color ? `--card-accent:${esc(item.color)};` : "";
  const iconHTML    = buildIconHTML(item.icon);
  const statusHTML  = item.statusCheck
    ? `<span class="status-dot unknown" data-id="${esc(item.id)}" data-tooltip="Checking…" role="status" aria-label="status unknown"></span>`
    : "";
  const descHTML = item.description
    ? `<p class="card-desc" title="${esc(item.description)}">${esc(item.description)}</p>`
    : "";
  const tagsHTML = (item.tags || []).length
    ? `<div class="card-tags">${item.tags.map((t) => `<span class="tag">${esc(t)}</span>`).join("")}</div>`
    : "";

  return `
    <a class="card"
       href="${esc(item.url)}"
       target="${esc(item.target || "_blank")}"
       rel="noopener noreferrer"
       data-id="${esc(item.id)}"
       data-title="${esc(item.title).toLowerCase()}"
       data-desc="${esc(item.description || "").toLowerCase()}"
       data-tags="${esc((item.tags || []).join(" ").toLowerCase())}"
       style="${accentStyle}"
       title="${esc(item.title)}${item.url ? ` – ${esc(item.url)}` : ""}">
      <div class="card-top">
        <div class="card-icon">${iconHTML}</div>
        ${statusHTML}
      </div>
      <div class="card-title">${esc(item.title)}</div>
      ${descHTML}
      ${tagsHTML}
    </a>`;
}

function buildIconHTML(icon) {
  if (!icon) return "🔗";
  if (icon.startsWith("http://") || icon.startsWith("https://") || icon.startsWith("/")) {
    return `<img src="${esc(icon)}" alt="" loading="lazy" />`;
  }
  return esc(icon);
}

// ── Status updates ────────────────────────────────────────────────────────────
function updateStatusDots() {
  $$(".status-dot[data-id]").forEach((dot) => {
    const status = statuses[dot.dataset.id];
    if (!status) return;

    dot.classList.remove("up", "slow", "down", "unknown");
    const cls = dotClass(status);
    dot.classList.add(cls);
    dot.dataset.tooltip = dotLabel(status);
    dot.setAttribute("aria-label", `status ${cls}`);
  });
}

function dotClass(s) {
  if (!s.up) return "down";
  if (s.responseTimeMs > 1000) return "slow";
  return "up";
}

function dotLabel(s) {
  if (!s.up) return `Offline${s.error ? `: ${s.error}` : ""}`;
  return `Online · ${s.responseTimeMs}ms`;
}

// ── Search / filter ───────────────────────────────────────────────────────────
function applySearch(raw) {
  const q = raw.trim().toLowerCase();
  state.query = q;

  let totalVisible = 0;

  $$(".section").forEach((sec) => {
    let sectionVisible = 0;
    sec.querySelectorAll(".card").forEach((card) => {
      const match =
        !q ||
        card.dataset.title.includes(q) ||
        card.dataset.desc.includes(q) ||
        card.dataset.tags.includes(q);
      card.hidden = !match;
      if (match) sectionVisible++;
    });
    sec.hidden = sectionVisible === 0;
    totalVisible += sectionVisible;
  });

  el.empty().hidden = totalVisible > 0 || !q;
  if (q) el.emptyQuery().textContent = q;
}

// ── Polling ───────────────────────────────────────────────────────────────────
function startPolling() {
  const interval = Math.max((config.checkInterval || 30), 5) * 1000;
  setInterval(loadStatus, interval);
}

// ── Refresh ───────────────────────────────────────────────────────────────────
async function triggerRefresh() {
  const btn = el.btnRefresh();
  btn.classList.add("spinning");
  try {
    await fetch("/api/reload", { method: "POST" });
    await sleep(800);
    await loadConfig();
    renderDashboard();
    await loadStatus();
    showToast("✓ Dashboard refreshed");
  } catch {
    showToast("⚠ Refresh failed");
  } finally {
    btn.classList.remove("spinning");
  }
}

// ── Toast ─────────────────────────────────────────────────────────────────────
let toastTimer = null;
function showToast(msg, ms = 3000) {
  const t = el.toast();
  t.textContent = msg;
  t.classList.add("visible");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => t.classList.remove("visible"), ms);
}

// ── Event bindings ────────────────────────────────────────────────────────────
function bindEvents() {
  el.search().addEventListener("input", (e) => applySearch(e.target.value));
  el.btnRefresh().addEventListener("click", triggerRefresh);

  document.addEventListener("keydown", (e) => {
    // "/" or Ctrl+K → focus search
    if ((e.key === "/" || (e.ctrlKey && e.key === "k")) && document.activeElement !== el.search()) {
      e.preventDefault();
      el.search().focus();
      el.search().select();
      return;
    }
    // Esc → clear search
    if (e.key === "Escape" && document.activeElement === el.search()) {
      el.search().value = "";
      applySearch("");
      el.search().blur();
      return;
    }
    // R → refresh (when not typing)
    if (e.key === "r" && !isTyping()) {
      triggerRefresh();
    }
  });
}

// ── Helpers ───────────────────────────────────────────────────────────────────
function esc(str) {
  if (typeof str !== "string") return "";
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function isTyping() {
  const tag = document.activeElement?.tagName;
  return tag === "INPUT" || tag === "TEXTAREA" || document.activeElement?.isContentEditable;
}

function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}
