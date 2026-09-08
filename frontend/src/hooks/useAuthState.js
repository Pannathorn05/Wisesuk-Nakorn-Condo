import { useEffect, useState } from "react";
import { isLoggedIn, onAuthChange } from "../utils/authStorage";

// รู้ทันทีว่า login/logout เปลี่ยนสถานะ (ใช้ใน Header — FE-24) ไม่ต้อง reload หน้าเพื่อเห็นผล
export function useAuthState() {
  const [loggedIn, setLoggedIn] = useState(isLoggedIn);

  useEffect(() => onAuthChange(() => setLoggedIn(isLoggedIn())), []);

  return loggedIn;
}
