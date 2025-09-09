"use client"
import { useState } from "react";
import { useRouter } from "next/navigation"; // ← hna
import styles from "./Login.module.css";

export default function Login() {
  const router = useRouter();
  const [form, setForm] = useState({ email: "", password: "" });
  const [err, setErr] = useState("");

  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const res = await fetch("http://localhost:8080/api/login", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(form),
    });
    if (!res.ok) return setErr(await res.text());
    router.push("/");
  }

  return (
    <div className={styles.container}>
      <div className={styles.loginCard}>
        <h1 className={styles.title}>Login</h1>
        <form className={styles.form} onSubmit={submit}>
          <input
            className={styles.input}
            placeholder="email"
            value={form.email}
            onChange={(e) => setForm({ ...form, email: e.target.value })}
          />
          <input
            className={styles.input}
            type="password"
            placeholder="password"
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
          />
          <div className={styles.buttonGroup}>
            <button className={styles.loginButton} type="submit">Login</button>
            <button
              type="button"
              className={styles.registerButton}
              onClick={() => router.push("/register")}
            >
              Register
            </button>
          </div>
        </form>
        {err && <p className={styles.error}>{err}</p>}
      </div>
    </div>
  );
}

