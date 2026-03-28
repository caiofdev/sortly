import useFileOrganizerController from './controllers/useFileOrganizerController';
import useLanguagePreference from './hooks/useLanguagePreference';
import useOrganizationOptions from './hooks/useOrganizationOptions';
import OrganizerView from './views/OrganizerView';

function App() {
  const { language, setLanguage } = useLanguagePreference();
  const { organizationOptions, updateOrganizationOption } = useOrganizationOptions();

  const {
    sourceFolderPath,
    destinationFolderPath,
    hasUndo,
    isLoading,
    loadingAction,
    feedback,
    handleClearFeedback,
    handleResolveDroppedPath,
    handleSelectSourceFolder,
    handleSelectDestinationFolder,
    handleOrganizeFiles,
    handleUndoLastOrganization
  } = useFileOrganizerController(language, organizationOptions);

  return (
    <OrganizerView
      language={language}
      organizationOptions={organizationOptions}
      sourceFolderPath={sourceFolderPath}
      destinationFolderPath={destinationFolderPath}
      hasUndo={hasUndo}
      isLoading={isLoading}
      loadingAction={loadingAction}
      feedback={feedback}
      onClearFeedback={handleClearFeedback}
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