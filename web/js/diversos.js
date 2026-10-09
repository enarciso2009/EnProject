function atualizarBarraVisual(input) {
    let valor = parseInt(input.value) || 0;
    if (valor > 100) { valor = 100; input.value = 100; }
    if (valor < 0) { valor = 0; input.value = 0; }
    
    // Atualiza a largura da barra verde
    const barra = input.parentElement.querySelector(".progress-bar-fill");
    if (barra) barra.style.width = valor + "%";

    // Se chegar a 100%, marca o checkbox de concluído automaticamente
    const row = input.closest(".tarefa-row");
    const checkbox = row.querySelector('input[type="checkbox"]');
    const hidden = row.querySelector('input[name="tarefa_concluido[]"]');
    
    if (valor === 100) {
        if (checkbox) checkbox.checked = true;
        if (hidden) hidden.value = "true";
    } else {
        if (checkbox) checkbox.checked = false;
        if (hidden) hidden.value = "false";
    }
}

function vincularConclusao(checkbox) {
    const row = checkbox.closest(".tarefa-row");
    const hidden = row.querySelector('input[name="tarefa_concluido[]"]');
    const progressoInput = row.querySelector('input[name="tarefa_progresso[]"]');
    const barra = row.querySelector(".progress-bar-fill");

    if (checkbox.checked) {
        if (hidden) hidden.value = "true";
        if (progressoInput) progressoInput.value = 100;
        if (barra) barra.style.width = "100%";
    } else {
        if (hidden) hidden.value = "false";
        if (progressoInput && progressoInput.value == 100) {
            progressoInput.value = 90; // Regressa para 90% se desmarcar
            if (barra) barra.style.width = "90%";
        }
    }
}



function adicionarTarefa() {
    const container = document.getElementById("tarefas-container");
    const row = document.createElement("div");
    row.className = "tarefa-row";
    row.innerHTML = `
        <input type="text" name="tarefa_nome[]" placeholder="Nova Tarefa" required>
        <input type="text" name="tarefa_resp[]" placeholder="Responsável" required>
        <input type="date" name="tarefa_inicio[]" required>
        <input type="date" name="tarefa_fim[]" required>
        <div class="progress-container">
            <input type="number" name="tarefa_progresso[]" value="0" min="0" max="100" placeholder="%" required oninput="atualizarBarraVisual(this)">
            <div class="progress-bar-bg"><div class="progress-bar-fill" style="width: 0%;"></div></div>
        </div>
        <input type="checkbox" onchange="vincularConclusao(this)">
        <input type="hidden" name="tarefa_concluido[]" value="false">
        <button type="button" class="btn-del" onclick="removerLinha(this)">X</button>
    `;
    container.appendChild(row);
}

function removerLinha(botao) {
    botao.parentElement.remove();
}


// Exibe miniaturas das imagens selecionadas localmente
function visualizarImagens(input) {
    const container = document.getElementById("preview-imagens");
    container.innerHTML = "";
    if (input.files) {
        Array.from(input.files).forEach(file => {
            const reader = new FileReader();
            reader.onload = function(e) {
                const img = document.createElement("img");
                img.className = "preview-img";
                img.src = e.target.result;
                container.appendChild(img);
            }
            reader.readAsDataURL(file);
        });
    }
}

function excluirFotoExistente(botao) {
    botao.parentElement.remove();
}