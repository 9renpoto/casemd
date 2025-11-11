/**
 * @typedef {Object} DefaultState
 * @property {string} name
 * @property {string} markdown
 * @property {string} csv
 */

/**
 * @typedef {Object} PreviewResponse
 * @property {string} [csv]
 */

/**
 * Extracts the default state from the DOM.
 *
 * @returns {DefaultState}
 */
function loadDefaultState() {
  /** @type {HTMLScriptElement | null} */
  const script = document.getElementById("default-state");
  if (!script) {
    return { name: "", markdown: "", csv: "" };
  }

  try {
    /** @type {DefaultState} */
    const defaults = JSON.parse(script.textContent ?? "{}");
    return {
      name: defaults.name ?? "",
      markdown: defaults.markdown ?? "",
      csv: defaults.csv ?? "",
    };
  } catch (error) {
    console.warn("Failed to parse default state", error);
    return { name: "", markdown: "", csv: "" };
  }
}

/**
 * Registers event handlers for the preview form.
 *
 * @param {HTMLFormElement} form
 * @param {HTMLButtonElement} submitButton
 * @param {HTMLElement} statusElement
 * @param {HTMLElement} errorElement
 * @param {HTMLTableSectionElement} headElement
 * @param {HTMLTableSectionElement} bodyElement
 * @param {DefaultState} defaults
 * @returns {void}
 */
function initializePreview(
  form,
  submitButton,
  statusElement,
  errorElement,
  headElement,
  bodyElement,
  defaults,
) {
  renderTable(defaults.csv, headElement, bodyElement);

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const nameField = form.querySelector("#name");
    const markdownField = form.querySelector("#markdown");

    const name = nameField instanceof HTMLInputElement ? nameField.value : "";
    const markdown = markdownField instanceof HTMLTextAreaElement ? markdownField.value : "";

    submitButton.disabled = true;
    statusElement.textContent = "Converting...";
    errorElement.textContent = "";

    try {
      const response = await fetch("/api/preview", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, markdown }),
      });

      if (!response.ok) {
        throw new Error((await response.text()) || "conversion failed");
      }

      /** @type {PreviewResponse} */
      const result = await response.json();
      renderTable(result.csv ?? "", headElement, bodyElement);
      statusElement.textContent = "Done";
    } catch (error) {
      errorElement.textContent = error instanceof Error ? error.message : String(error);
      statusElement.textContent = "";
      renderTable("", headElement, bodyElement);
    } finally {
      submitButton.disabled = false;
    }
  });
}

/**
 * Renders CSV content inside the provided table sections.
 *
 * @param {string} csvText
 * @param {HTMLTableSectionElement} headElement
 * @param {HTMLTableSectionElement} bodyElement
 * @returns {void}
 */
function renderTable(csvText, headElement, bodyElement) {
  headElement.innerHTML = "";
  bodyElement.innerHTML = "";

  if (!csvText.trim()) {
    bodyElement.innerHTML = '<tr><td class="placeholder">No data available.</td></tr>';
    return;
  }

  const rows = parseCSV(csvText);
  if (!rows.length) {
    bodyElement.innerHTML = '<tr><td class="placeholder">No data available.</td></tr>';
    return;
  }

  const [header, ...dataRows] = rows;

  const headerRow = document.createElement("tr");
  header.forEach((cell) => {
    const th = document.createElement("th");
    th.textContent = cell;
    headerRow.appendChild(th);
  });
  headElement.appendChild(headerRow);

  if (!dataRows.length) {
    const row = document.createElement("tr");
    const cell = document.createElement("td");
    cell.className = "placeholder";
    cell.textContent = "No rows parsed.";
    row.appendChild(cell);
    bodyElement.appendChild(row);
    return;
  }

  dataRows.forEach((values) => {
    const row = document.createElement("tr");
    header.forEach((_, index) => {
      const cell = document.createElement("td");
      cell.textContent = values[index] ?? "";
      row.appendChild(cell);
    });
    bodyElement.appendChild(row);
  });
}

/**
 * Converts CSV text into a matrix of values.
 *
 * @param {string} text
 * @returns {string[][]}
 */
function parseCSV(text) {
  /** @type {string[][]} */
  const rows = [];
  /** @type {string[]} */
  let current = [];
  let value = "";
  let inQuotes = false;

  for (let index = 0; index < text.length; index += 1) {
    const char = text[index];

    if (inQuotes) {
      if (char === '"') {
        if (text[index + 1] === '"') {
          value += '"';
          index += 1;
        } else {
          inQuotes = false;
        }
      } else {
        value += char;
      }
      continue;
    }

    switch (char) {
      case '"':
        inQuotes = true;
        break;
      case ',':
        current.push(value);
        value = "";
        break;
      case '\r':
        break;
      case '\n':
        current.push(value);
        rows.push(current);
        current = [];
        value = "";
        break;
      default:
        value += char;
    }
  }

  if (value !== "" || current.length) {
    current.push(value);
    rows.push(current);
  }

  return rows;
}

(function bootstrap() {
  const defaults = loadDefaultState();

  const form = document.getElementById("preview-form");
  const submitButton = document.getElementById("submit");
  const statusElement = document.getElementById("status");
  const errorElement = document.getElementById("error");
  const headElement = document.getElementById("csv-head");
  const bodyElement = document.getElementById("csv-body");

  if (
    !(form instanceof HTMLFormElement) ||
    !(submitButton instanceof HTMLButtonElement) ||
    !(statusElement instanceof HTMLElement) ||
    !(errorElement instanceof HTMLElement) ||
    !(headElement instanceof HTMLTableSectionElement) ||
    !(bodyElement instanceof HTMLTableSectionElement)
  ) {
    console.warn("Preview UI is missing expected elements.");
    return;
  }

  initializePreview(
    form,
    submitButton,
    statusElement,
    errorElement,
    headElement,
    bodyElement,
    defaults,
  );
})();
