// Original illustrations inspired by smiski & sonny angel figures — not official artwork.

type Props = { className?: string };

/** A little glow-in-the-dark green buddy hugging its knees and reading. */
export function SmiBuddy({ className = "" }: Props) {
  return (
    <svg viewBox="0 0 120 130" className={`glow ${className}`} aria-hidden="true">
      {/* body, sitting with knees up */}
      <path
        d="M30 78 C26 104 34 122 60 122 C86 122 94 104 90 78 Z"
        fill="var(--smi-body)"
        stroke="var(--smi-shade)"
        strokeWidth="2"
      />
      {/* head: slightly flat-topped, like the real figures */}
      <path
        d="M60 14 C86 14 98 30 98 52 C98 74 82 86 60 86 C38 86 22 74 22 52 C22 30 34 14 60 14 Z"
        fill="var(--smi-body)"
        stroke="var(--smi-shade)"
        strokeWidth="2"
      />
      {/* eyes */}
      <circle cx="47" cy="54" r="3" fill="#2b3324" />
      <circle cx="73" cy="54" r="3" fill="#2b3324" />
      {/* book */}
      <g>
        <path d="M34 92 L60 98 L60 120 L34 114 Z" fill="#fff7ea" stroke="#d9b98f" strokeWidth="1.5" />
        <path d="M86 92 L60 98 L60 120 L86 114 Z" fill="#fff2df" stroke="#d9b98f" strokeWidth="1.5" />
        <line x1="40" y1="100" x2="54" y2="103" stroke="#e3cba8" strokeWidth="1.2" />
        <line x1="40" y1="105" x2="54" y2="108" stroke="#e3cba8" strokeWidth="1.2" />
        <line x1="66" y1="103" x2="80" y2="100" stroke="#e3cba8" strokeWidth="1.2" />
      </g>
      {/* little hands holding the book */}
      <ellipse cx="33" cy="102" rx="6" ry="7" fill="var(--smi-body)" stroke="var(--smi-shade)" strokeWidth="2" />
      <ellipse cx="87" cy="102" rx="6" ry="7" fill="var(--smi-body)" stroke="var(--smi-shade)" strokeWidth="2" />
    </svg>
  );
}

/** A cherub baby in a bunny-eared hood, resting on break. */
export function AngelBuddy({ className = "" }: Props) {
  return (
    <svg viewBox="0 0 120 130" className={className} aria-hidden="true">
      {/* wings */}
      <path d="M30 92 C10 86 8 104 22 108 C14 114 26 122 36 110 Z" fill="#ffffff" stroke="#e8dfe0" strokeWidth="1.5" />
      <path d="M90 92 C110 86 112 104 98 108 C106 114 94 122 84 110 Z" fill="#ffffff" stroke="#e8dfe0" strokeWidth="1.5" />
      {/* body */}
      <path d="M38 88 C34 110 42 124 60 124 C78 124 86 110 82 88 Z" fill="var(--skin)" stroke="#e8bfa8" strokeWidth="1.5" />
      {/* bunny ears */}
      <path d="M40 34 C30 4 44 -2 50 26 Z" fill="var(--blush)" stroke="#e39aa7" strokeWidth="1.5" />
      <path d="M80 34 C90 4 76 -2 70 26 Z" fill="var(--blush)" stroke="#e39aa7" strokeWidth="1.5" />
      <path d="M42 30 C37 14 43 10 47 26 Z" fill="#fff" opacity="0.6" />
      <path d="M78 30 C83 14 77 10 73 26 Z" fill="#fff" opacity="0.6" />
      {/* hood */}
      <circle cx="60" cy="58" r="36" fill="var(--blush)" stroke="#e39aa7" strokeWidth="1.5" />
      {/* face */}
      <circle cx="60" cy="62" r="27" fill="var(--skin)" />
      {/* sleepy happy eyes */}
      <path d="M46 61 Q50 57 54 61" stroke="#3a2b2b" strokeWidth="2.2" fill="none" strokeLinecap="round" />
      <path d="M66 61 Q70 57 74 61" stroke="#3a2b2b" strokeWidth="2.2" fill="none" strokeLinecap="round" />
      {/* cheeks + mouth */}
      <ellipse cx="44" cy="70" rx="5" ry="3" fill="#f6a8b4" opacity="0.7" />
      <ellipse cx="76" cy="70" rx="5" ry="3" fill="#f6a8b4" opacity="0.7" />
      <path d="M57 72 Q60 75 63 72" stroke="#c97a80" strokeWidth="1.8" fill="none" strokeLinecap="round" />
    </svg>
  );
}
