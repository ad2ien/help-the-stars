// Filter the table based on user input
function filterTable() {
  const input = document.getElementById("filter-input");
  const filter = input.value.toLowerCase();
  const table = document.getElementById("repos-table");
  const rows = table.getElementsByTagName("tr");

  for (let i = 1; i < rows.length; i++) {
    // Start from 1 to skip the header row
    const repoCell = rows[i].getElementsByTagName("td")[0]; // Repository column
    const descriptionCell = rows[i].getElementsByTagName("td")[1]; // description column
    const languageCell = rows[i].getElementsByTagName("td")[2]; // language
    if (repoCell || descriptionCell) {
      const repoText = repoCell.textContent.toLowerCase();
      const issueTitleText = descriptionCell.textContent.toLowerCase();
      const languageText = languageCell.textContent.toLowerCase();
      if (
        repoText.includes(filter) ||
        issueTitleText.includes(filter) ||
        languageText.includes(filter)
      ) {
        rows[i].style.display = "";
      } else {
        rows[i].style.display = "none";
      }
    }
  }
}
