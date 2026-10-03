import { useEffect, useState } from 'react';

const identity = (value) => value;

function readStorage(key) {
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

function writeStorage(key, value) {
  try {
    window.localStorage.setItem(key, value);
  } catch {
    // Armazenamento indisponível (modo privado, cota): a preferência só não é salva.
  }
}

// Estado salvo no localStorage. deserialize recebe o texto salvo (ou null) e
// devolve o valor inicial; serialize converte o valor em texto para salvar.
// Falhas de leitura ou gravação nunca quebram a interface.
function usePersistentState(key, { deserialize, serialize = identity }) {
  const [value, setValue] = useState(() => deserialize(readStorage(key)));

  useEffect(() => {
    writeStorage(key, serialize(value));
  }, [key, serialize, value]);

  return [value, setValue];
}

export default usePersistentState;
