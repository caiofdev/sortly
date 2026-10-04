import { DEFAULT_LANGUAGE, SUPPORTED_LANGUAGES } from '../i18n/language';
import usePersistentState from './usePersistentState';

const LANGUAGE_STORAGE_KEY = 'sortly.language';

const deserializeLanguage = (saved) =>
  SUPPORTED_LANGUAGES.includes(saved) ? saved : DEFAULT_LANGUAGE;

function useLanguagePreference() {
  const [language, setLanguage] = usePersistentState(LANGUAGE_STORAGE_KEY, {
    deserialize: deserializeLanguage
  });

  return {
    language,
    setLanguage
  };
}

export default useLanguagePreference;
