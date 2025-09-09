"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

type User = {
  id: number;
  username: string;
  email: string;
  role: "player" | "watcher";
};

export default function Profile() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    fetch("http://localhost:8080/api/me", { credentials: "include" })
      .then(async (r) => {
        if (!r.ok) {
          router.replace("/login"); // redirect if not authenticated
          return null;
        }
        return r.json();
      })
      .then((data) => {
        if (data) setUser(data);
      })
      .finally(() => setLoading(false));
  }, [router]);

  async function logout(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const res = await fetch("http://localhost:8080/api/logout", {
      method: "POST",
      credentials: "include",
    });
    if (!res.ok) {
     // alert(await res.text());
      return;
    }
    router.replace("/login"); // redirect to login page
  }

  if (loading) return <p>Loading...</p>;
  if (!user) return null;

  return (
    <div>
      <h1>Welcome {user.username}</h1>
      <p>Role: {user.role}</p>
      <form onSubmit={logout}>
        <button type="submit">Logout</button>
      </form>
    </div>
  );
}
