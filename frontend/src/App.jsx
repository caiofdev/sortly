import useFileOrganizerController from './controllers/useFileOrganizerController';
import useNotifications from './hooks/useNotifications';
import useSettings from './hooks/useSettings';
import OrganizerView from './views/OrganizerView';

function App() {
  const { settings, setLanguage, setCriterion } = useSettings();
  const { language, criteria } = settings;
  const { notifications, notify, clearNotifications } = useNotifications(language);

  const {
    sourceFolderPath,
    destinationFolderPath,
    hasUndo,
    isLoading,
    loadingAction,
    handleSelectSourceFolder,
    handleSelectDestinationFolder,
    handleOrganizeFiles,
    handleUndoLastOrganization,
    handleLanguageChange,
    handleCriterionChange
  } = useFileOrganizerController({ language, notify, setLanguage, setCriterion });

  return (
    <OrganizerView
      language={language}
      criteria={criteria}
      sourceFolderPath={sourceFolderPath}
      destinationFolderPath={destinationFolderPath}
      hasUndo={hasUndo}
      isLoading={isLoading}
      loadingAction={loadingAction}
      notifications={notifications}
      onClearNotifications={clearNotifications}
      onLanguageChange={handleLanguageChange}
      onCriterionChange={handleCriterionChange}
      onSelectSourceFolder={handleSelectSourceFolder}
      onSelectDestinationFolder={handleSelectDestinationFolder}
      onOrganizeFiles={handleOrganizeFiles}
      onUndoLastOrganization={handleUndoLastOrganization}
    />
  );
}

export default App;
