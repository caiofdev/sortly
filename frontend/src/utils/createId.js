let fallbackCounter = 0;

// Identificador único para itens de lista (ex.: notificações).
export function createId() {
  if (globalThis.crypto?.randomUUID) {
    return globalThis.crypto.randomUUID();
  }
  fallbackCounter += 1;
  return `${Date.now()}-${fallbackCounter}`;
}
