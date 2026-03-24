import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { AUTH_REFRESH_COOKIE, AUTH_ROLE_COOKIE } from "@/lib/constants";
import { getRoleHome } from "@/lib/navigation/role-home";

export default async function HomePage() {
  const cookieStore = await cookies();
  const refreshToken = cookieStore.get(AUTH_REFRESH_COOKIE)?.value;
  const role = cookieStore.get(AUTH_ROLE_COOKIE)?.value;

  if (!refreshToken) {
    redirect("/login");
  }

  redirect(getRoleHome(role));
}
