function confirmarExclusao(botao) {
    const projetoId = botao.getAttribute('data-id');
    let cliquesRestantes = parseInt(botao.getAttribute('data-cliques'));

    cliquesRestantes--;

    // Configura mensagens personalizadas para cada um dos primeiros cliques
    if (cliquesRestantes === 2) {
        if (!confirm("⚠️ Atenção: Você tem certeza de que deseja excluir este projeto? (Confirmação 1 de 3)")) {
            resetarBotao(botao);
            return; // Cancela se o usuário clicar em "Cancelar"
        }
        botao.setAttribute('data-cliques', cliquesRestantes);
        botao.innerText = "Confirmar (2/3)";
        botao.style.backgroundColor = "#ffc107";
        botao.style.color = "#000";
    } 
    else if (cliquesRestantes === 1) {
        if (!confirm("🚨 Cuidado: Todos os dados, tarefas e imagens serão apagados permanentemente! Tem certeza absoluta? (Confirmação 2 de 3)")) {
            resetarBotao(botao);
            return;
        }
        botao.setAttribute('data-cliques', cliquesRestantes);
        botao.innerText = "Último clique (3/3)";
        botao.style.backgroundColor = "#fd7e14"; // Laranja de perigo iminente
        botao.style.color = "#fff";
    } 
    else {
        // Terceiro e último clique: executa a ação no servidor
        botao.innerText = "Excluindo...";
        botao.disabled = true;
    
        fetch(`/projeto/excluir?id=${projetoId}`, {
            method: 'DELETE'
        })
        .then(response => {
            if (response.ok) {
                alert("🚀 Projeto excluído com sucesso!");
                window.location.reload();
            } else {
                alert("❌ Erro ao excluir o projeto no servidor.");
                resetarBotao(botao);
            }
        })
        .catch(error => {
            console.error("Erro:", error);
            alert("🌐 Erro de conexão ao tentar excluir.");
            resetarBotao(botao);
        });
    }
}

function resetarBotao(botao) {
    botao.setAttribute('data-cliques', '3');
    botao.innerText = "Excluir";
    botao.style.backgroundColor = "#dc3545";
    botao.style.color = "#ffffff";
    botao.disabled = false;
}