
  document.querySelector("form").addEventListener("submit", async function(event) {
    event.preventDefault(); // Prevent the default form submission

    const form = event.target;
    const formData = new FormData(form);

    try {
        const response = await fetch("/api/add/post", {
            method: "POST",
            body: formData
        });

        if (!response.ok) {
            throw new Error("Network response was not ok " + response.statusText);
        }

        const result = await response.text();
        console.log("Form submitted successfully:", result);
        alert("Form submitted successfully!");
    } catch (error) {
        console.error("There was a problem with the fetch operation:", error);
    }
}); 
