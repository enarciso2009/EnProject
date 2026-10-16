function adicionarTarefa() {
    console.log("-> Adicionando uma nova linha de tarefa...");

    const container = document.getElementById("tarefas-container");
    if (!container) {
        console.error("Erro: Container de tarefas não encontrado.");
        return;
    }

    // Cria a nova linha estruturada
    const novaLinha = document.createElement("div");
    novaLinha.className = "tarefa-row";

    // 🌟 HTML ATUALIZADO: Inclui a caixinha progress-container com a faixa verde de progressão
    novaLinha.innerHTML = `
        <input type="text" name="tarefa_nome[]" placeholder="Ex: Ajustar API" required>
        <input type="text" name="tarefa_resp[]" placeholder="Nome do dev" required>
        <input type="date" name="tarefa_inicio[]" required>
        <input type="date" name="tarefa_fim[]" required>
        
        <!-- Bloco de Progresso com a Faixa Verde de Progressão -->
        <div class="progress-container">
            <input type="number" name="tarefa_progresso[]" min="0" max="100" placeholder="%" required oninput="atualizarBarraVisual(this)">
            <div class="progress-bar-bg" style="background: #e9ecef; border-radius: 4px; width: 60px; height: 8px; overflow: hidden; margin-top: 4px;">
                <div class="progress-bar-fill" style="width: 0%; background: #28a745; height: 100%; transition: width 0.2s;"></div>
            </div>
        </div>

        <input type="checkbox" onchange="vincularConclusao(this)">
        <input type="hidden" name="tarefa_concluido[]" value="false">
        <button type="button" class="btn-del" onclick="removerLinha(this)">X</button>
    `;

    // Insere a nova tarefa no final da lista
    container.appendChild(novaLinha);
    console.log("-> Nova linha inserida com sucesso!");
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

// Atualiza a largura da barrinha verde dinamicamente ao digitar a porcentagem
function atualizarBarraVisual(inputNumero) {
    let valor = parseInt(inputNumero.value) || 0;
    if (valor < 0) valor = 0;
    if (valor > 100) valor = 100;
    inputNumero.value = valor;

    // Localiza a div da barra de preenchimento logo abaixo do input
    const row = inputNumero.closest('.progress-container');
    const fill = row.querySelector('.progress-bar-fill');
    if (fill) {
        fill.style.width = valor + '%';
    }
}

// Vincula o clique do checkbox com o preenchimento automático do progresso em 100%
function vincularConclusao(checkbox) {
    const row = checkbox.closest('.tarefa-row');
    const inputHidden = row.querySelector('input[name="tarefa_concluido[]"]');
    const inputProgresso = row.querySelector('input[name="tarefa_progresso[]"]');

    if (checkbox.checked) {
        if (inputHidden) inputHidden.value = "true";
        if (inputProgresso) {
            inputProgresso.value = 100;
            atualizarBarraVisual(inputProgresso);
        }
    } else {
        if (inputHidden) inputHidden.value = "false";
        if (inputProgresso && inputProgresso.value == 100) {
            inputProgresso.value = 0;
            atualizarBarraVisual(inputProgresso);
        }
    }
}
