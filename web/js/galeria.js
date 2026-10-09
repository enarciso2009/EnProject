var url = "{{ . }}";
var extensao = url.split('.').pop().toLowerCase();
if (['jpg', 'jpeg', 'png', 'gif', 'webp'].includes(extensao)) {
    document.write('<img src="{{ . }}" class="preview-img" alt="Foto do Projeto">');
} else {
    document.write('<div style="width: 100px; height: 100px; display: flex; align-items: center; justify-content: center; background: #e9ecef; border-radius: 6px; font-weight: bold; color: #495057; font-size: 12px; border: 1px solid #ccc;">📄 .' + extensao.toUpperCase() + '</div>');
}