# ARIA Widget Patterns

Code snippets for common ARIA widget patterns. Reference these when implementing interactive components.

## Dialog / Modal

```tsx
function Dialog({ open, onClose, title, children }) {
  const dialogRef = useRef(null);
  const triggerRef = useRef(null);

  useEffect(() => {
    if (open) {
      // Save trigger and focus first focusable element
      triggerRef.current = document.activeElement;
      dialogRef.current?.focus();
    }
    return () => {
      // Return focus to trigger on close
      triggerRef.current?.focus();
    };
  }, [open]);

  if (!open) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="dialog-title"
      ref={dialogRef}
      tabIndex={-1}
      onKeyDown={(e) => {
        if (e.key === "Escape") onClose();
        // Trap focus: wrap Tab/Shift+Tab within dialog
      }}
    >
      <h2 id="dialog-title">{title}</h2>
      {children}
      <button onClick={onClose}>Close</button>
    </div>
  );
}
```

**Key requirements:**
- `role="dialog"` + `aria-modal="true"`
- `aria-labelledby` points to the dialog title
- Focus moves into dialog on open
- Focus returns to trigger element on close
- Escape key closes the dialog
- Focus is trapped inside (Tab/Shift+Tab cycle within)

## Tabs

```tsx
function Tabs({ tabs }) {
  const [activeIndex, setActiveIndex] = useState(0);

  const handleKeyDown = (e, index) => {
    let newIndex = index;
    if (e.key === "ArrowRight") newIndex = (index + 1) % tabs.length;
    if (e.key === "ArrowLeft") newIndex = (index - 1 + tabs.length) % tabs.length;
    if (e.key === "Home") newIndex = 0;
    if (e.key === "End") newIndex = tabs.length - 1;

    if (newIndex !== index) {
      e.preventDefault();
      setActiveIndex(newIndex);
    }
  };

  return (
    <div>
      <div role="tablist" aria-label="Content tabs">
        {tabs.map((tab, i) => (
          <button
            key={tab.id}
            role="tab"
            id={`tab-${tab.id}`}
            aria-selected={i === activeIndex}
            aria-controls={`panel-${tab.id}`}
            tabIndex={i === activeIndex ? 0 : -1}
            onKeyDown={(e) => handleKeyDown(e, i)}
            onClick={() => setActiveIndex(i)}
          >
            {tab.label}
          </button>
        ))}
      </div>
      {tabs.map((tab, i) => (
        <div
          key={tab.id}
          role="tabpanel"
          id={`panel-${tab.id}`}
          aria-labelledby={`tab-${tab.id}`}
          hidden={i !== activeIndex}
          tabIndex={0}
        >
          {tab.content}
        </div>
      ))}
    </div>
  );
}
```

**Key requirements:**
- Arrow keys navigate between tabs (roving tabindex)
- Home/End jump to first/last tab
- Only active tab is in tab order (`tabindex="0"`)
- `aria-selected` indicates active tab
- `aria-controls` / `aria-labelledby` link tabs to panels

## Accordion / Disclosure

```tsx
function Accordion({ items }) {
  const [openItems, setOpenItems] = useState(new Set());

  const toggle = (id) => {
    setOpenItems((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  };

  return (
    <div>
      {items.map((item) => (
        <div key={item.id}>
          <h3>
            <button
              aria-expanded={openItems.has(item.id)}
              aria-controls={`content-${item.id}`}
              onClick={() => toggle(item.id)}
            >
              {item.title}
            </button>
          </h3>
          <div
            id={`content-${item.id}`}
            role="region"
            aria-labelledby={`heading-${item.id}`}
            hidden={!openItems.has(item.id)}
          >
            {item.content}
          </div>
        </div>
      ))}
    </div>
  );
}
```

**Key requirements:**
- Trigger is a `<button>` inside a heading
- `aria-expanded` reflects open/closed state
- `aria-controls` links button to content panel
- Content panel has `role="region"` with `aria-labelledby`
- Consider native `<details>` / `<summary>` for simple cases

## Combobox (Autocomplete)

```tsx
function Combobox({ options, label, onSelect }) {
  const [query, setQuery] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const filtered = options.filter((o) =>
    o.label.toLowerCase().includes(query.toLowerCase())
  );

  const handleKeyDown = (e) => {
    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        setActiveIndex((i) => Math.min(i + 1, filtered.length - 1));
        setIsOpen(true);
        break;
      case "ArrowUp":
        e.preventDefault();
        setActiveIndex((i) => Math.max(i - 1, 0));
        break;
      case "Enter":
        if (activeIndex >= 0) {
          onSelect(filtered[activeIndex]);
          setIsOpen(false);
        }
        break;
      case "Escape":
        setIsOpen(false);
        setActiveIndex(-1);
        break;
    }
  };

  return (
    <div>
      <label id="combo-label">{label}</label>
      <input
        role="combobox"
        aria-expanded={isOpen}
        aria-controls="combo-listbox"
        aria-labelledby="combo-label"
        aria-activedescendant={
          activeIndex >= 0 ? `option-${filtered[activeIndex].id}` : undefined
        }
        aria-autocomplete="list"
        value={query}
        onChange={(e) => {
          setQuery(e.target.value);
          setIsOpen(true);
          setActiveIndex(-1);
        }}
        onKeyDown={handleKeyDown}
      />
      {isOpen && (
        <ul id="combo-listbox" role="listbox">
          {filtered.map((option, i) => (
            <li
              key={option.id}
              id={`option-${option.id}`}
              role="option"
              aria-selected={i === activeIndex}
              onClick={() => {
                onSelect(option);
                setIsOpen(false);
              }}
            >
              {option.label}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
```

**Key requirements:**
- `role="combobox"` on the input
- `aria-expanded`, `aria-controls`, `aria-activedescendant`
- Arrow keys navigate options, Enter selects, Escape closes
- `role="listbox"` on the dropdown, `role="option"` on each item

## Alert / Live Region

```tsx
// Assertive — announced immediately (errors, urgent messages)
function ErrorAlert({ message }) {
  return (
    <div role="alert" className="text-destructive">
      {message}
    </div>
  );
}

// Polite — announced at next pause (status updates)
function StatusMessage({ message }) {
  return (
    <div role="status" aria-live="polite">
      {message}
    </div>
  );
}

// Dynamic — render the container first, then update content
// The live region must exist in the DOM BEFORE content changes
function SearchResults({ count }) {
  return (
    <div aria-live="polite" aria-atomic="true">
      {count} results found
    </div>
  );
}
```

**Key requirements:**
- `role="alert"` = `aria-live="assertive"` (use for errors)
- `role="status"` = `aria-live="polite"` (use for updates)
- The live region container must be in the DOM before content changes
- Use `aria-atomic="true"` when the entire region should be re-announced
