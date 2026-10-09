function renumerarItens() {
    const linhas = document.querySelectorAll("#materiais-container .mat-row");
    linhas.forEach((linha, index) => {
        linha.querySelector('input[name="mat_item[]"]').value = index + 1;
    });
}

function adicionarMaterial() {
    const container = document.getElementById("materiais-container");
    const row = document.createElement("div");
    row.className = "tarefa-row mat-row";
    row.style.gridTemplateColumns = "1fr 5fr 2fr 1fr";
    row.innerHTML = `
        <input type="number" name="mat_item[]" readonly style="background: #f0f0f0; text-align: center;">
        <input type="text" name="mat_descricao[]" placeholder="Descrição do Material" required>
        <input type="number" name="mat_quantidade[]" value="1" min="1" required>
        <button type="button" class="btn-del" onclick="removerMaterial(this)">X</button>
    `;
    container.appendChild(row);
    renumerarItens();
}

function removerMaterial(botao) {
    botao.parentElement.remove();
    renumerarItens();
}