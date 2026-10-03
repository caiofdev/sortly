import { useCallback } from 'react';
import { DEFAULT_OPTIONS, normalizeOptions, toggleOption } from '../domain/organizationOptions';
import usePersistentState from './usePersistentState';

const STORAGE_KEY = 'sortly.organizationOptions';

function deserializeOptions(saved) {
  if (!saved) {
    return DEFAULT_OPTIONS;
  }
  try {
    return normalizeOptions(JSON.parse(saved));
  } catch {
    return DEFAULT_OPTIONS;
  }
}

function useOrganizationOptions() {
  const [organizationOptions, setOrganizationOptions] = usePersistentState(STORAGE_KEY, {
    deserialize: deserializeOptions,
    serialize: JSON.stringify
  });

  const updateOrganizationOption = useCallback(
    (key, value) => setOrganizationOptions((current) => toggleOption(current, key, value)),
    [setOrganizationOptions]
  );

  return {
    organizationOptions,
    updateOrganizationOption
  };
}

export default useOrganizationOptions;
