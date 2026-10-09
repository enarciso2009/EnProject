async function enviarRelatorioDiario() {
    const dataInput = document.getElementById("diario_data");
    const descricaoInput = document.getElementById("diario_descricao");
    const imagensInput = document.getElementById("imagens_diario");
    const participantesInput = document.getElementById("diario_participantes");
    const pendenciasInput = document.getElementById("diario_pendencias");
    const veiculosInput = document.getElementById("diario_veiculos");
    
    const projetoIdInput = document.querySelector('input[name="id"]');
    const projetoId = projetoIdInput ? projetoIdInput.value : "";

    if (!dataInput.value || !descricaoInput.value.trim()) {
        alert("Por favor, preencha a Data, Descrição e Participantes do relatório diário antes de salvar.");
        return;
    }

    const formData = new FormData();
    formData.append("projeto_id", projetoId);
    formData.append("data", dataInput.value);
    formData.append("descricao", descricaoInput.value);
    formData.append("participantes", participantesInput.value);
    formData.append("pendencias", pendenciasInput.value);
    formData.append("veiculos", veiculosInput.value);
    if (imagensInput.files) {
        Array.from(imagensInput.files).forEach(file => {
            formData.append("imagens_diario[]", file);
        });
    }

    try {
        const resposta = await fetch("/projetos/diario", {
            method: "POST",
            body: formData
        });

        if (!resposta.ok) {
            throw new Error("Erro do servidor ao salvar o relatório.");
        }

        const resultado = await resposta.json();

        const msgVazio = document.getElementById("msg-vazio");
        if (msgVazio) msgVazio.remove();

        const dataFormatada = dataInput.value.split('-').reverse().join('/');

        let galeriaHtml = '';
        if (resultado.imagens && resultado.imagens.length > 0) {
            galeriaHtml = `<div class="historico-galeria">`;
            resultado.imagens.forEach(url => {
                galeriaHtml += `
                    <a href="${url}" target="_blank">
                        <img src="${url}" class="historico-thumb" alt="Evidência Visual">
                    </a>`;
            });
            galeriaHtml += `</div>`;
        }

        const novoItem = document.createElement("details");
        novoItem.className = "historico-item";
        novoItem.open = true;
        novoItem.style.borderLeftColor = "#17a2b8";
        novoItem.innerHTML = `
            <summary class="historico-header">
                <span class="historico-data">${dataFormatada}</span>
                <span class="seta-indicadora">Clique para fechar </span>
            </summary>
            <div class="conteudo-expandido">
                <div class="historico-texto">${descricaoInput.value}</div>
                <div class="historico-texto">${pendenciasInput.value}</div>
                <div class="historico-texto">${participantesInput.value}</div>
                <div class="historico-texto">${veiculosInput.value}</div>
                ${galeriaHtml}
            </div>
        `;

        const lista = document.getElementById("historico-lista");
        lista.insertBefore(novoItem, lista.firstChild);

        dataInput.value = "";
        descricaoInput.value = "";
        pendenciasInput.value = "";
        participantesInput.value = "";
        veiculosInput.value = "";   
        imagensInput.value = "";
        document.getElementById("preview-imagens-novas-diario").innerHTML = "";

        alert("Relatório Diário registrado com sucesso!");

    } catch (error) {
        console.error(error);
        alert("Falha ao registrar o relatório. Verifique seu backend.");
    }
}


function prepararEdicaoDiario(botao) {
    // 1. O JavaScript sobe até o container do histórico e lê a div oculta de forma segura
    const container = botao.closest('.conteudo-expandido');
    const dados = container.querySelector('.dados-ocultos');
    
    const id = dados.getAttribute('data-id');
    const descricao = dados.getAttribute('data-desc');
    const pendencias = dados.getAttribute('data-pend');
    const participantes = dados.getAttribute('data-part');
    const veiculos = dados.getAttribute('data-veic');

    // 2. Mapeia os campos de texto do formulário na sua página
    const campoDesc =  document.getElementById('diario-descricao');
    const campoPend =  document.getElementById('diario-pendencias');
    const campoPart = document.getElementById('diario-participantes');
    const campoVeic = document.getElementById('diario-veiculos');
    
    if (campoDesc) campoDesc.value = descricao;
    if (campoPend) campoPend.value = pendencias;
    if (campoPart) campoPart.value = participantes;
    if (campoVeic) campoVeic.value = veiculos;

    // 3. Insere o ID oculto no formulário para o Go saber que é um UPDATE
    let formDiario = campoDesc ? campoDesc.closest('form') : null;
    if (formDiario) {
        let inputId = formDiario.querySelector('input[name="diario_id"]');
        if (!inputId) {
            inputId = document.createElement('input');
            inputId.type = 'hidden';
            inputId.name = 'diario_id';
            formDiario.appendChild(inputId);
        }
        inputId.value = id;
        
        // Atualiza o texto do botão do formulário
        const btnSalvar = formDiario.querySelector('button[type="submit"]');
        if (btnSalvar) btnSalvar.innerText = "💾 Salvar Alterações do Relatório";
        
        // Rola a tela até o formulário para você começar a editar
        formDiario.scrollIntoView({ behavior: 'smooth' });
    }
}

function excluirRelatorio(botao) {
    // Aqui está o seu código de exclusão atualizado para ler o ID de forma segura:
    const container = botao.closest('.conteudo-expandido');
    const dados = container.querySelector('.dados-ocultos');
    const relatorioId = dados.getAttribute('data-id');

    if (confirm("⚠️ Tem certeza que deseja apagar permanentemente este relatório diário?")) {
        fetch(`/projetos/diario/excluir?id=${relatorioId}`, {
            method: 'DELETE'
        })
        .then(response => {
            if (response.ok) {
                alert("✨ Relatório excluído com sucesso!");
                window.location.reload(); 
            } else {
                alert("❌ Erro ao tentar excluir o relatório.");
            }
        })
        .catch(error => {
            console.error("Erro:", error);
            alert("🌐 Erro de conexão.");
        });
    }
}


