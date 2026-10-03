import useFileOrganizerController from './controllers/useFileOrganizerController';
import useLanguagePreference from './hooks/useLanguagePreference';
import useNotifications from './hooks/useNotifications';
import useOrganizationOptions from './hooks/useOrganizationOptions';
import OrganizerView from './views/OrganizerView';

function App() {
  const { language, setLanguage } = useLanguagePreference();
  const { organizationOptions, updateOrganizationOption } = useOrganizationOptions();
  const { notifications, notify, clearNotifications } = useNotifications(language);

  const {
    sourceFolderPath,
    destinationFolderPath,
    hasUndo,
    isLoading,
    loadingAction,
    handleResolveDroppedPath,
    handleSelectSourceFolder,
    handleSelectDestinationFolder,
    handleOrganizeFiles,
    handleUndoLastOrganization
  } = useFileOrganizerController({ language, organizationOptions, notify });

  return (
    <OrganizerView
      language={language}
      organizationOptions={organizationOptions}
      sourceFolderPath={sourceFolderPath}
      destinationFolderPath={destinationFolderPath}
      hasUndo={hasUndo}
      isLoading={isLoading}
      loadingAction={loadingAction}
      notifications={notifications}
      onClearNotifications={clearNotifications}
      onLanguageChange={setLanguage}
      onOptionChange={updateOrganizationOption}
      onResolveDroppedPath={handleResolveDroppedPath}
      onSelectSourceFolder={handleSelectSourceFolder}
      onSelectDestinationFolder={handleSelectDestinationFolder}
      onOrganizeFiles={handleOrganizeFiles}
      onUndoLastOrganization={handleUndoLastOrganization}
    />
  );
}

export default App;
