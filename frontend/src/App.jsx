import { BrowserRouter } from "react-router-dom";
import { AppRoutes } from "./routes/AppRoutes";

// Header/Footer ย้ายไปอยู่ใน layout ของแต่ละโซนแล้ว (GuestLayout / AdminAuthLayout)
// เพราะโซนผู้ดูแลใช้ header คนละแบบและไม่มี footer — ดู docs/task/frontend/06-admin-login.md (FE-26)
function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  );
}

export default App;
