(() => {
  function renderResult(result) {
    if (result.value === null) {
      return `${result.kind}\n\nnull`;
    }

    if (typeof result.value === "object") {
      return `${result.kind} in ${result.duration_ms}ms\n\n${JSON.stringify(result.value, null, 2)}`;
    }

    return `${result.kind} in ${result.duration_ms}ms\n\n${String(result.value)}`;
  }

  function attachRunner(panel) {
    const button = panel.querySelector("[data-run-button]");
    const output = panel.querySelector("[data-run-output]");
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

  document.addEventListener("DOMContentLoaded", () => {
    document.querySelectorAll("[data-example-runner]").forEach(attachRunner);
  });
})();
