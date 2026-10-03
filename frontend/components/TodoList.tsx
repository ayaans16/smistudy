"use client";

import { useEffect, useRef, useState } from "react";
import { todos as todoApi, type Todo } from "@/lib/api";
import { errorMessage } from "./AuthCard";
import { SmiBuddy } from "./Mascots";

// Same order as the API: open items oldest first, then completed ones most recent first.
function sortTodos(list: Todo[]) {
  return [...list].sort((a, b) => {
    if (a.done !== b.done) return a.done ? 1 : -1;
    if (a.done) return (b.doneAt ?? "").localeCompare(a.doneAt ?? "");
    return a.createdAt.localeCompare(b.createdAt);
  });
}

export default function TodoList() {
  const [items, setItems] = useState<Todo[] | null>(null);
  const [text, setText] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    todoApi
      .list()
      .then((list) => !cancelled && setItems(list))
      .catch((e) => !cancelled && setError(errorMessage(e)));
    return () => {
      cancelled = true;
    };
  }, []);

  // Apply a change instantly, then reconcile with the server (or roll back on failure).
  async function mutate(optimistic: (list: Todo[]) => Todo[], request: () => Promise<Todo[] | Todo | void>) {
    const before = items ?? [];
    setItems(sortTodos(optimistic(before)));
    setError(null);
    try {
      const result = await request();
      if (result && !Array.isArray(result)) {
        setItems((list) => sortTodos((list ?? []).map((t) => (t.id === result.id ? result : t))));
      }
    } catch (e) {
      setItems(before);
      setError(errorMessage(e));
    }
  }

  async function add(e: React.FormEvent) {
    e.preventDefault();
    const value = text.trim();
    if (!value) return;
    setText("");
    setError(null);
    try {
      const created = await todoApi.add(value);
      setItems((list) => sortTodos([...(list ?? []), created]));
    } catch (err) {
      setText(value);
      setError(errorMessage(err));
    }
  }

  const toggle = (t: Todo) =>
    mutate(
      (list) => list.map((x) => (x.id === t.id ? { ...x, done: !t.done, doneAt: t.done ? undefined : new Date().toISOString() } : x)),
      () => todoApi.update(t.id, { done: !t.done }),
    );
  const rename = (t: Todo, value: string) =>
    mutate((list) => list.map((x) => (x.id === t.id ? { ...x, text: value } : x)), () => todoApi.update(t.id, { text: value }));
  const remove = (t: Todo) => mutate((list) => list.filter((x) => x.id !== t.id), () => todoApi.remove(t.id));
  const clearDone = () => mutate((list) => list.filter((x) => !x.done), () => todoApi.clearDone().then(() => undefined));

  const doneCount = items?.filter((t) => t.done).length ?? 0;
  const total = items?.length ?? 0;
  const allDone = total > 0 && doneCount === total;

  return (
    <section className="rounded-3xl border border-line bg-card p-6 shadow-sm">
      <div className="flex items-baseline justify-between">
        <h2 className="text-lg font-bold">To-do</h2>
        {total > 0 && (
          <span className="text-sm font-semibold text-muted">
            {doneCount} of {total} done
          </span>
        )}
      </div>

      <form onSubmit={add} className="mt-4 flex gap-2">
        <input
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Add a task, then press Enter"
          maxLength={200}
          aria-label="New to-do"
          className="min-w-0 flex-1 rounded-full border border-line bg-bg px-4 py-2 text-sm outline-none transition focus:border-accent"
        />
        <button
          disabled={!text.trim()}
          className="rounded-full bg-accent px-4 py-2 text-sm font-bold text-[#1f2a14] transition hover:brightness-105 disabled:opacity-40"
        >
          Add
        </button>
      </form>

      {error && <p className="mt-3 text-sm text-red-500">{error}</p>}

      {items && total === 0 && (
        <div className="flex items-center gap-3 py-5 text-sm text-muted">
          <SmiBuddy className="h-10 w-10 shrink-0" />
          Nothing on your list yet. What are you studying today?
        </div>
      )}

      {allDone && <p className="mt-4 text-center text-sm font-bold text-accent-strong">All done, nice work! 🌱</p>}

      <ul className="mt-3 space-y-1">
        {items?.map((t) => (
          <TodoItem key={t.id} todo={t} onToggle={() => toggle(t)} onRename={(v) => rename(t, v)} onDelete={() => remove(t)} />
        ))}
      </ul>

      {doneCount > 0 && (
        <button onClick={clearDone} className="mt-3 text-xs font-semibold text-muted transition hover:text-fg">
          Clear completed ({doneCount})
        </button>
      )}
    </section>
  );
}

function TodoItem({
  todo,
  onToggle,
  onRename,
  onDelete,
}: {
  todo: Todo;
  onToggle: () => void;
  onRename: (text: string) => void;
  onDelete: () => void;
}) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(todo.text);
  const input = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editing) input.current?.select();
  }, [editing]);

  function startEditing() {
    setDraft(todo.text);
    setEditing(true);
  }

  function save() {
    setEditing(false);
    const value = draft.trim();
    if (value && value !== todo.text) onRename(value);
  }

  return (
    <li className="group flex items-center gap-3 rounded-xl px-2 py-1.5 transition hover:bg-bg">
      <button
        role="checkbox"
        aria-checked={todo.done}
        aria-label={todo.done ? `Mark "${todo.text}" as not done` : `Mark "${todo.text}" as done`}
        onClick={onToggle}
        className={`flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 transition ${
          todo.done ? "border-accent bg-accent text-[#1f2a14]" : "border-line hover:border-accent"
        }`}
      >
        {todo.done && (
          <svg viewBox="0 0 16 16" className="h-3 w-3" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <path d="M3.5 8.5l3 3 6-7" />
          </svg>
        )}
      </button>

      {editing ? (
        <input
          ref={input}
          value={draft}
          maxLength={200}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={save}
          onKeyDown={(e) => {
            if (e.key === "Enter") save();
            if (e.key === "Escape") setEditing(false);
          }}
          aria-label="Edit to-do"
          className="min-w-0 flex-1 rounded-lg border border-accent bg-card px-2 py-0.5 text-sm outline-none"
        />
      ) : (
        <span
          onDoubleClick={startEditing}
          title="Double-click to edit"
          className={`min-w-0 flex-1 break-words text-sm transition ${todo.done ? "text-muted line-through decoration-2" : ""}`}
        >
          {todo.text}
        </span>
      )}

      <button
        onClick={onDelete}
        aria-label={`Delete "${todo.text}"`}
        className="shrink-0 rounded-md px-1.5 text-muted opacity-100 transition hover:text-red-500 sm:opacity-0 sm:group-hover:opacity-100 sm:focus:opacity-100"
      >
        ×
      </button>
    </li>
  );
}
