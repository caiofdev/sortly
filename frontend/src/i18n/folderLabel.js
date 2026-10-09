// O backend diz se as pastas de 1º nível são categorias do critério Tipo
// (categoryFolders); só então o nome neutro (images) vira o rótulo traduzido.
// Assim, uma pasta de extensão chamada "audio" não é traduzida por engano (#81).
export function folderLabel(name, categoryFolders, labels) {
  if (!name) return labels.previewRoot;
  if (!categoryFolders) return name;
  return labels.categories[name] ?? name;
}

// O progresso traz a pasta relativa ao destino (images\2026-09-14); só o 1º nível
// pode ser categoria, e o resto do caminho fica como veio (#84).
export function folderPathLabel(path, categoryFolders, labels) {
  const [, first, rest] = /^([^\\/]*)(.*)$/s.exec(path);
  return folderLabel(first, categoryFolders, labels) + rest;
}
