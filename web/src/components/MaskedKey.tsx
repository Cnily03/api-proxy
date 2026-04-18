import { useState } from "react";

import { Copyable } from "@/components/CopyButton";

function maskKey(key: string): string {
  if (!key) return "-";
  if (key.length > 13) return `${key.slice(0, 10)}***`;
  return `${key.slice(0, Math.max(0, key.length - 3))}***`;
}

export function EyeToggle({ visible, onToggle }: { visible: boolean; onToggle: () => void }) {
  return (
    <button
      type="button"
      className="inline-flex items-center justify-center w-4 h-4 text-muted hover:text-foreground cursor-pointer transition-colors"
      onClick={onToggle}
      aria-label={visible ? "隐藏" : "显示"}
    >
      <span className={`w-3.5 h-3.5 ${visible ? "icon-[mdi--eye-off]" : "icon-[mdi--eye]"}`} />
    </button>
  );
}

export function KeyRow({ label, value }: { label: string; value: string }) {
  const [visible, setVisible] = useState(false);
  return (
    <>
      <span className="text-muted inline-flex items-center gap-1.5">
        <span>{label}</span>
        {value ? <EyeToggle visible={visible} onToggle={() => setVisible((v) => !v)} /> : null}
      </span>
      <Copyable text={value}>
        <span className="text-foreground break-all">{!value ? "-" : visible ? value : maskKey(value)}</span>
      </Copyable>
    </>
  );
}

export { maskKey };
