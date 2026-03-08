(() => {
  function formatDuration(us) {
    if (us < 1000) return `${us}µs`;
    return `${(us / 1000).toFixed(1)}ms`;
  }

  function renderResult(result) {
    const dur = formatDuration(result.duration_us);
    if (result.value === null) {
      return `${result.kind}\n\nnull`;
    }

    if (typeof result.value === "object") {
      return `${result.kind} in ${dur}\n\n${JSON.stringify(result.value, null, 2)}`;
    }

    return `${result.kind} in ${dur}\n\n${String(result.value)}`;
  }

  function attachRunner() {
    const button = document.querySelector("[data-run-button]");
    const output = document.querySelector("[data-run-output]");
    if (!button || !output) {
      return;
    }

    button.addEventListener("click", async () => {
      const originalLabel = button.textContent;
      button.disabled = true;
      button.textContent = "Running...";
      output.textContent = "Executing run() through the site runner...";

      try {
        const response = await fetch(button.dataset.runUrl, {
          method: "POST",
          headers: {
            "Accept": "application/json",
          },
        });
        const payload = await response.json();
        if (!response.ok) {
          output.textContent = payload.error || "Execution failed.";
          return;
        }

        output.textContent = renderResult(payload.result);
      } catch (error) {
        output.textContent = error instanceof Error ? error.message : "Execution failed.";
      } finally {
        button.disabled = false;
        button.textContent = originalLabel;
      }
    });
  }

  function escapeHtml(str) {
    return str
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  function highlightVibescript(source) {
    const lines = source.split("\n");
    const result = [];

    const keywords = new Set([
      "def", "end", "if", "elsif", "else", "enum", "return",
      "while", "for", "in", "do", "class", "module", "require",
      "and", "or", "not", "then", "unless", "until", "case",
      "when", "break", "next", "yield", "begin", "rescue",
      "ensure", "raise", "import", "export", "let", "const",
      "var", "fn", "match", "struct", "impl", "trait", "pub",
      "self", "super", "new", "puts", "print", "println",
    ]);
    const constants = new Set(["true", "false", "nil"]);

    for (const line of lines) {
      const tokens = [];
      let i = 0;

      while (i < line.length) {
        // Comment
        if (line[i] === "#") {
          tokens.push(`<span class="tok-comment">${escapeHtml(line.slice(i))}</span>`);
          i = line.length;
          continue;
        }

        // String
        if (line[i] === '"') {
          let j = i + 1;
          while (j < line.length && line[j] !== '"') {
            if (line[j] === "\\") j++;
            j++;
          }
          j = Math.min(j + 1, line.length);
          tokens.push(`<span class="tok-string">${escapeHtml(line.slice(i, j))}</span>`);
          i = j;
          continue;
        }

        // Number
        if (/\d/.test(line[i]) && (i === 0 || /[^a-zA-Z_]/.test(line[i - 1]))) {
          let j = i;
          while (j < line.length && /[\d.]/.test(line[j])) j++;
          if (j > i && !/[a-zA-Z_]/.test(line[j] || "")) {
            tokens.push(`<span class="tok-number">${line.slice(i, j)}</span>`);
            i = j;
            continue;
          }
        }

        // Word (identifiers, keywords, constants)
        if (/[a-zA-Z_]/.test(line[i])) {
          let j = i;
          while (j < line.length && /[a-zA-Z0-9_]/.test(line[j])) j++;
          const word = line.slice(i, j);

          if (constants.has(word)) {
            tokens.push(`<span class="tok-constant">${word}</span>`);
          } else if (keywords.has(word)) {
            if (word === "def") {
              // Look ahead for function name
              let k = j;
              while (k < line.length && line[k] === " ") k++;
              let nameStart = k;
              while (k < line.length && /[a-zA-Z0-9_]/.test(line[k])) k++;
              if (k > nameStart) {
                tokens.push(`<span class="tok-keyword">def</span>`);
                tokens.push(escapeHtml(line.slice(j, nameStart)));
                tokens.push(`<span class="tok-function">${escapeHtml(line.slice(nameStart, k))}</span>`);
                i = k;
                continue;
              }
            }
            tokens.push(`<span class="tok-keyword">${word}</span>`);
          } else if (/^[A-Z]/.test(word)) {
            // Check for Enum::Variant
            if (line.slice(j, j + 2) === "::") {
              let k = j + 2;
              let vs = k;
              while (k < line.length && /[a-zA-Z0-9_]/.test(line[k])) k++;
              tokens.push(`<span class="tok-type">${escapeHtml(word)}</span>::<span class="tok-type">${escapeHtml(line.slice(vs, k))}</span>`);
              i = k;
              continue;
            }
            tokens.push(`<span class="tok-type">${escapeHtml(word)}</span>`);
          } else {
            // Check if followed by ( → method/function call
            let k = j;
            while (k < line.length && line[k] === " ") k++;
            if (line[k] === "(" && i > 0 && line[i - 1] === ".") {
              tokens.push(`<span class="tok-function">${escapeHtml(word)}</span>`);
            } else {
              tokens.push(escapeHtml(word));
            }
          }
          i = j;
          continue;
        }

        // Operators
        const twoChar = line.slice(i, i + 2);
        if (["->", "==", "!=", "<=", ">=", "&&", "||"].includes(twoChar)) {
          tokens.push(`<span class="tok-operator">${escapeHtml(twoChar)}</span>`);
          i += 2;
          continue;
        }
        if ("=+-*/%<>!".includes(line[i])) {
          tokens.push(`<span class="tok-operator">${escapeHtml(line[i])}</span>`);
          i++;
          continue;
        }

        // Default: plain character
        tokens.push(escapeHtml(line[i]));
        i++;
      }

      result.push(tokens.join(""));
    }

    return result.join("\n");
  }

  function initThemeToggle() {
    const toggle = document.querySelector("[data-theme-toggle]");
    if (!toggle) return;

    toggle.addEventListener("click", () => {
      const current = document.documentElement.getAttribute("data-theme");
      const next = current === "dark" ? "light" : "dark";
      document.documentElement.setAttribute("data-theme", next);
      localStorage.setItem("theme", next);
    });
  }

  function initCatalog() {
    const grid = document.querySelector("[data-catalog-grid]");
    const nav = document.querySelector("[data-catalog-nav]");
    const filtersEl = document.querySelector("[data-active-filters]");
    if (!grid || !nav) return;

    const cards = Array.from(grid.querySelectorAll(".example-card"));
    const categories = {};
    cards.forEach((card) => {
      const cat = card.dataset.category || "Other";
      if (!categories[cat]) categories[cat] = [];
      categories[cat].push(card);
    });

    const sorted = Object.keys(categories).sort((a, b) => {
      if (a === "Vibescript Showcase") return -1;
      if (b === "Vibescript Showcase") return 1;
      return a.localeCompare(b);
    });

    let activeCategory = null;

    function render() {
      cards.forEach((card) => {
        const cat = card.dataset.category || "Other";
        const show = !activeCategory || cat === activeCategory;
        card.hidden = !show;
      });

      nav.querySelectorAll(".catalog-nav-item").forEach((btn) => {
        btn.classList.toggle("is-active", btn.dataset.cat === (activeCategory || "__all__"));
      });

      if (filtersEl) {
        filtersEl.innerHTML = "";
        if (activeCategory) {
          const pill = document.createElement("button");
          pill.className = "filter-pill";
          pill.innerHTML = `${escapeHtml(activeCategory)} <span class="filter-x">&times;</span>`;
          pill.addEventListener("click", () => {
            activeCategory = null;
            render();
          });
          filtersEl.appendChild(pill);
        }
      }
    }

    const allBtn = document.createElement("button");
    allBtn.className = "catalog-nav-item is-active";
    allBtn.dataset.cat = "__all__";
    allBtn.innerHTML = `<span>All</span><span class="catalog-nav-count">${cards.length}</span>`;
    allBtn.addEventListener("click", () => {
      activeCategory = null;
      render();
    });
    nav.appendChild(allBtn);

    sorted.forEach((cat) => {
      const btn = document.createElement("button");
      btn.className = "catalog-nav-item";
      btn.dataset.cat = cat;
      btn.innerHTML = `<span>${escapeHtml(cat)}</span><span class="catalog-nav-count">${categories[cat].length}</span>`;
      btn.addEventListener("click", () => {
        activeCategory = cat;
        render();
      });
      nav.appendChild(btn);
    });
  }

  function initExpandToggle() {
    const toggle = document.querySelector("[data-expand-toggle]");
    if (!toggle) return;

    toggle.addEventListener("click", () => {
      const grid = toggle.closest(".detail-grid");
      if (!grid) return;
      grid.classList.toggle("output-expanded");
    });
  }

  document.addEventListener("DOMContentLoaded", () => {
    attachRunner();
    initExpandToggle();

    document.querySelectorAll("code.language-vibescript").forEach((el) => {
      el.innerHTML = highlightVibescript(el.textContent);
    });

    initThemeToggle();
    initCatalog();
  });
})();
