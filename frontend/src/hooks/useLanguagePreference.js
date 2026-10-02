import { useEffect, useState } from 'react';

const LANGUAGE_STORAGE_KEY = 'sortly.language';
const SUPPORTED_LANGUAGES = ['pt-BR', 'en'];

function useLanguagePreference() {
  const [language, setLanguage] = useState(() => {
    const savedLanguage = window.localStorage.getItem(LANGUAGE_STORAGE_KEY);
    return SUPPORTED_LANGUAGES.includes(savedLanguage) ? savedLanguage : 'pt-BR';
  });

  useEffect(() => {
    window.localStorage.setItem(LANGUAGE_STORAGE_KEY, language);
  }, [language]);

  return {
    language,
    setLanguage
  };
}

export default useLanguagePreference;
