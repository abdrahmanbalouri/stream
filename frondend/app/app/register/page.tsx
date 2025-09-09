"use client"
import { useState } from "react";
import { useRouter } from "next/navigation"; // ← hna

export default function Register() {
  const router = useRouter(); // ← hna
  const [form, setForm] = useState({ username: "", email: "", password: "", role: "watcher" });
  const [err, setErr] = useState("");

  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const res = await fetch("http://localhost:8080/api/register", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(form),
    });
    if (!res.ok) return setErr(await res.text());
    router.push("/login"); // ← badal Router.push b router.push
  }

  return (
    <div>
      <h1>Register</h1>
      <form onSubmit={submit}>
        <input
          placeholder="username"
          value={form.username}
          onChange={e => setForm({ ...form, username: e.target.value })}
        />
        <input
          placeholder="email"
          value={form.email}
          onChange={e => setForm({ ...form, email: e.target.value })}
        />
        <input
          type="password"
          placeholder="password"
          value={form.password}
          onChange={e => setForm({ ...form, password: e.target.value })}
        />
        <select
          value={form.role}
          onChange={e => setForm({ ...form, role: e.target.value })}
        >
          <option value="watcher">Watcher</option>
          <option value="player">Player</option>
        </select>
        <button type="submit">Register</button>
      </form>
      {err && <p style={{ color: "red" }}>{err}</p>}
    </div>
  );
}
