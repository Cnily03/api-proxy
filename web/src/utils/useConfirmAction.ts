import { useCallback, useRef, useState } from "react";

/**
 * Double-click confirmation hook.
 * First click enters "pending" state; second click within timeout executes action.
 * After timeout, state resets automatically.
 */
export function useConfirmAction(action: () => void, timeout = 3000) {
  const [pending, setPending] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const reset = useCallback(() => {
    setPending(false);
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  const trigger = useCallback(() => {
    if (pending) {
      reset();
      action();
    } else {
      setPending(true);
      timerRef.current = setTimeout(() => {
        setPending(false);
        timerRef.current = null;
      }, timeout);
    }
  }, [pending, action, timeout, reset]);

  return { pending, trigger, reset };
}
