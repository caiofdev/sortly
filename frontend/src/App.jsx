import { useEffect, useLayoutEffect, useMemo } from 'react';
import useViewState from './hooks/useViewState';
import { toNotificationItems } from './i18n/notifications';
import OrganizerView from './views/OrganizerView';

function App() {
  const { state, ready, actions } = useViewState();
  const { language, theme, duplicates, criteria } = state.settings;
  const notifications = useMemo(
    () => toNotificationItems(state.notifications, language),
    [state.notifications, language]
  );

  // Leitores de tela escolhem a pronúncia pelo lang do documento (#74).
  useEffect(() => {
    document.documentElement.lang = language;
  }, [language]);

  // No <html>, para os tokens do tema valerem também fora do #root. Antes da
  // pintura: com useEffect, a primeira tela saía no tema escuro por um quadro
  // (ADR 0006, #76).
  useLayoutEffect(() => {
    document.documentElement.dataset.theme = theme;
  }, [theme]);

  if (!ready) {
    return null;
  }

  return (
    <OrganizerView
      language={language}
      theme={theme}
      duplicates={duplicates}
      criteria={criteria}
      sourceFolderPath={state.sourceFolderPath}
      destinationFolderPath={state.destinationFolderPath}
      preview={state.preview}
      progress={state.progress}
      lastResult={state.lastResult}
      history={state.history}
      hasUndo={state.hasUndo}
      isLoading={Boolean(state.busy)}
      loadingAction={state.busy || null}
      notifications={notifications}
      unread={state.unread}
      onClearNotifications={actions.clearNotifications}
      onMarkNotificationsRead={actions.markNotificationsRead}
      onLanguageChange={actions.setLanguage}
      onThemeChange={actions.setTheme}
      onDuplicatesChange={actions.setDuplicates}
      onCriterionChange={actions.setCriterion}
      onSelectSourceFolder={actions.selectSource}
      onSelectDestinationFolder={actions.selectDestination}
      onOrganizeFiles={actions.organize}
      onUndoLastOrganization={actions.undo}
      onCancel={actions.cancel}
      onOpenDestination={actions.openDestination}
      onStartOver={actions.startOver}
    />
  );
}

export default App;
