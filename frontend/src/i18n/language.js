export const DEFAULT_LANGUAGE = 'pt-BR';
export const SUPPORTED_LANGUAGES = ['pt-BR', 'en'];

// Textos do idioma pedido, ou do idioma padrão se ele não existir no dicionário.
export function getCopy(dictionary, language) {
  return dictionary[language] || dictionary[DEFAULT_LANGUAGE];
}

// Locale usado para formatar datas e horas.
export function toLocale(language) {
  return language === 'pt-BR' ? 'pt-BR' : 'en-US';
}
