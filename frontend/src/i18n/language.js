export const DEFAULT_LANGUAGE = 'pt-BR';
export const SUPPORTED_LANGUAGES = ['pt-BR', 'en'];

export function getCopy(dictionary, language) {
  return dictionary[language] || dictionary[DEFAULT_LANGUAGE];
}

export function toLocale(language) {
  return language === 'pt-BR' ? 'pt-BR' : 'en-US';
}
