import enIcon from '../assets/en.ico';
import ptBrIcon from '../assets/pt-br.ico';

const options = [
  { value: 'pt-BR', label: 'PT-BR', icon: ptBrIcon },
  { value: 'en', label: 'EN', icon: enIcon }
];

function LanguageToggle({ language, onChange }) {
  return (
    <div className="inline-flex rounded-2xl border border-white/10 bg-[#0F172A]/60 p-1">
      {options.map((option) => {
        const isActive = language === option.value;
        return (
          <button
            key={option.value}
            type="button"
            onClick={() => onChange(option.value)}
            className={`rounded-xl px-3 py-2 text-sm font-semibold transition-colors ${
              isActive ? 'bg-[#3B82F6] text-white' : 'text-[#94A3B8] hover:bg-white/5 hover:text-[#F8FAFC]'
            }`}
          >
            <img src={option.icon} alt={option.label} className="mr-2 inline-block h-5 w-5 rounded-full object-cover" />
            {option.label}
          </button>
        );
      })}
    </div>
  );
}

export default LanguageToggle;
