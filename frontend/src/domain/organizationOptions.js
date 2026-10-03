// Critérios de organização: chaves (na ordem dos checkboxes), valores padrão e
// a regra "pelo menos um critério marcado", num único lugar. O backend repete a
// validação (NO_CRITERIA) por segurança.

export const OPTION_KEYS = [
  'byDuration',
  'byPages',
  'byResolution',
  'byDate',
  'bySize',
  'byExtension'
];

export const DEFAULT_OPTIONS = Object.freeze({
  byDuration: false,
  byPages: false,
  byResolution: false,
  byDate: false,
  bySize: false,
  byExtension: true
});

export function countSelected(options) {
  return OPTION_KEYS.filter((key) => Boolean(options[key])).length;
}

// Só impede desmarcar o último critério marcado.
export function canToggle(options, key, value) {
  const isUncheckingLast = Boolean(options[key]) && !value && countSelected(options) === 1;
  return !isUncheckingLast;
}

export function toggleOption(options, key, value) {
  return canToggle(options, key, value) ? { ...options, [key]: value } : options;
}

// Completa um valor salvo com os padrões; byExtension ausente vale true.
export function normalizeOptions(saved) {
  return { ...DEFAULT_OPTIONS, ...saved, byExtension: saved?.byExtension ?? true };
}
