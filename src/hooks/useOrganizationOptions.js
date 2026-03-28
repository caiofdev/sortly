import { useEffect, useState } from 'react';

const STORAGE_KEY = 'sortly.organizationOptions';

const defaultOrganizationOptions = {
  byDuration: false,
  byPages: false,
  byResolution: false,
  byDate: false,
  bySize: false,
  byExtension: true
};

function useOrganizationOptions() {
  const [organizationOptions, setOrganizationOptions] = useState(() => {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return defaultOrganizationOptions;
    }

    try {
      const parsed = JSON.parse(raw);
      return {
        ...defaultOrganizationOptions,
        ...parsed,
        byExtension: parsed.byExtension ?? true
      };
    } catch {
      return defaultOrganizationOptions;
    }
  });

  useEffect(() => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(organizationOptions));
  }, [organizationOptions]);

  const updateOrganizationOption = (key, value) => {
    setOrganizationOptions((current) => ({
      ...current,
      [key]: value
    }));
  };

  return {
    organizationOptions,
    updateOrganizationOption
  };
}

export default useOrganizationOptions;
