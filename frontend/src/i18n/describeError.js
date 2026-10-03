import { UNEXPECTED } from '../services/sortlyGateway';

// Texto de erro para o usuário: tradução do código do backend; para erros sem
// código conhecido, a mensagem original (versão Electron) ou o texto padrão.
export function describeError(error, copy, fallback) {
  const translated = copy.errors[error?.code];
  if (translated) {
    return translated;
  }
  if (!error?.code || error.code === UNEXPECTED) {
    return legacyMessage(error) || fallback;
  }
  return fallback;
}

// Mensagens da versão Electron chegam como texto livre; códigos não traduzidos
// e erros vazios usam o texto padrão.
function legacyMessage(error) {
  const message = error?.message || '';
  return message === UNEXPECTED || message === 'Backend indisponível' ? '' : message;
}
