// O backend diz se as pastas de 1º nível são categorias do critério Tipo
// (categoryFolders); só então o nome neutro (images) vira o rótulo traduzido.
// Assim, uma pasta de extensão chamada "audio" não é traduzida por engano (#81).
export function folderLabel(name, categoryFolders, labels) {
  if (!name) return labels.previewRoot;
  if (!categoryFolders) return name;
  return labels.categories[name] ?? name;
}
