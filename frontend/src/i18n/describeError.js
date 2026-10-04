// Texto de erro para o usuário: tradução do código do backend, ou o texto
// padrão da ação para UNEXPECTED, códigos sem tradução e erros vazios.
export function describeError(error, copy, fallback) {
  return copy.errors[error?.code] || fallback;
}
