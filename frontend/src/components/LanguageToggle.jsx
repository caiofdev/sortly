import SegmentedControl from './SegmentedControl';

const options = [
  { value: 'pt-BR', label: 'PT-BR' },
  { value: 'en', label: 'EN' }
];

function LanguageToggle({ language, label, onChange }) {
  return <SegmentedControl label={label} options={options} value={language} onChange={onChange} />;
}

export default LanguageToggle;
