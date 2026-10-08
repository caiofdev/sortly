import { useEffect, useMemo } from 'react';
import useViewState from './hooks/useViewState';
import { toNotificationItems } from './i18n/notifications';
import OrganizerView from './views/OrganizerView';

function App() {
  const { state, ready, actions } = useViewState();
  const { language, criteria } = state.settings;
  const notifications = useMemo(
    () => toNotificationItems(state.notifications, language),
    [state.notifications, language]
  );

  // Leitores de tela escolhem a pronúncia pelo lang do documento (#74).
  useEffect(() => {
    document.documentElement.lang = language;
  }, [language]);

  if (!ready) {
    return null;
  }

  return (
    <OrganizerView
      language={language}
      criteria={criteria}
      sourceFolderPath={state.sourceFolderPath}
      destinationFolderPath={state.destinationFolderPath}
      hasUndo={state.hasUndo}
      isLoading={Boolean(state.busy)}
      loadingAction={state.busy || null}
      notifications={notifications}
      onClearNotifications={actions.clearNotifications}
      onLanguageChange={actions.setLanguage}
      onCriterionChange={actions.setCriterion}
      onSelectSourceFolder={actions.selectSource}
      onSelectDestinationFolder={actions.selectDestination}
      onOrganizeFiles={actions.organize}
      onUndoLastOrganization={actions.undo}
    />
  );
}

export default App;
