import { customAlphabet } from "nanoid";

const nanokey = customAlphabet("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 48);

export function generateKey() {
  return `sk-icu-${nanokey()}`;
}
