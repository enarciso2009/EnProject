async function enviarRelatorioDiario() {
    const dataInput = document.getElementById("diario_data");
    const descricaoInput = document.getElementById("diario_descricao");
    const imagensInput = document.getElementById("imagens_diario");
    const participantesInput = document.getElementById("diario_participantes");
    const pendenciasInput = document.getElementById("diario_pendencias");
    const veiculosInput = document.getElementById("diario_veiculos");
    
    const projetoIdInput = document.querySelector('input[name="id"]');
    const projetoId = projetoIdInput ? projetoIdInput.value : "";

    // 🌟 CAPTURA O ID OCULTO SE O BLOCO ESTIVER NO MODO EDIÇÃO
    const blocoDiario = document.getElementById('bloco-diario');
    const inputId = blocoDiario ? blocoDiario.querySelector('input[name="diario_id"]') : null;
    const isEdicao = inputId && inputId.value !== "";

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
    
    // 🌟 SE FOR EDIÇÃO, COMPARTILHA O ID COM O GO PARA EVITAR DUPLICIDADE
    if (isEdicao) {
        formData.append("diario_id", inputId.value);
    }

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

        // 🌟 SE FOR UMA EDIÇÃO CONCLUÍDA, ATUALIZA A TELA DO BANCO PARA EVITAR DUPLICADOS VISUAIS
        if (isEdicao) {
            alert("Relatório Diário alterado com sucesso!");
            window.location.reload();
            return;
        }

        // --- MANTÉM SEU COMPORTAMENTO ORIGINAL DE INSERÇÃO RÁPIDA APENAS PARA NOVOS CADASTROS ---
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

        // Limpa os campos após registrar novo item
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

function prepararEdicaoDiarioDefinitivo(botao) {
    console.log("-> Botão de editar relatório clicado!");
    
    try {
        // 1. Sobe até o bloco do histórico expandido
        const container = botao.closest('.conteudo-expandido');
        if (!container) {
            console.error("Erro: Não foi possível encontrar o bloco '.conteudo-expandido'.");
            return;
        }

        // 2. Localiza a div de dados ocultos que você já tem no HTML
        const dados = container.querySelector('.dados-ocultos');
        if (!dados) {
            console.error("Erro: Elemento '.dados-ocultos' não encontrado no HTML.");
            return;
        }

        // 3. Puxa os dados com segurança direto dos atributos data-*
        const id = dados.getAttribute('data-id');
        const dataRelatorio = dados.getAttribute('data-data');
        const descricao = dados.getAttribute('data-desc') || "";
        const pendencias = dados.getAttribute('data-pend') || "";
        const participantes = dados.getAttribute('data-part') || "";
        const veiculos = dados.getAttribute('data-veic') || "";

        console.log(`-> Dados Extraídos da Tag Oculta: ID=${id} | Veículos=${veiculos}`);

        // 4. Mapeia as caixas de texto do seu formulário na página (Lançar Novo Relatório Diário)
        const campoData = document.getElementById('diario_data');
        const campoDesc = document.getElementById('diario_descricao');
        const campoPend = document.getElementById('diario_pendencias');
        const campoPart = document.getElementById('diario_participantes');
        const campoVeic = document.getElementById('diario_veiculos');

        // Alerta caso os IDs das caixas de texto tenham mudado de nome na página
        if (!campoData || !campoDesc || !campoPend || !campoPart || !campoVeic) {
            alert("Erro Técnico: As caixas do formulário de digitação do diário não foram encontradas na página.");
            return;
        }

        // 5. Injeta os dados históricos de volta no formulário
        campoData.value = dataRelatorio;
        campoDesc.value = descricao;
        campoPend.value = pendencias;
        campoPart.value = participantes;
        campoVeic.value = veiculos;

        // 6. Gerencia o ID oculto no formulário para sinalizar Modo Edição ao Go
        const blocoDiario = document.getElementById('bloco-diario');
        if (blocoDiario) {
            let inputId = blocoDiario.querySelector('input[name="diario_id"]');  
            if (!inputId) { 
                inputId = document.createElement('input');
                inputId.type = 'hidden';
                inputId.name = 'diario_id';
                blocoDiario.appendChild(inputId);
            }    
            inputId.value = id;
            
            // Atualiza o texto do botão principal do formulário para dar o feedback de edição
            const btnSalvar = blocoDiario.querySelector('button');
            if (btnSalvar) {
                btnSalvar.innerText = "💾 Salvar Alterações do Relatório";
                btnSalvar.style.background = "#fd7e14"; // Muda para laranja indicando alteração
            }
            
            // Rola a tela suavemente até o formulário
            blocoDiario.scrollIntoView({ behavior: 'smooth' });
            console.log("-> Sucesso: Campos populados e tela movida!");
        }

    } catch (erro) {
        console.error("Falha ao processar o script JavaScript de edição:", erro);
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


