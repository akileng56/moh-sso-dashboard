import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import Dashboard from "./pages/dashboard/dashboard.component";

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<HomeRedirect />} />
        <Route path="/dashboard" element={<Dashboard />} />
      </Routes>
    </Router>
  );
}

function HomeRedirect() {
  window.location.href = "http://localhost:9000/api/v1/auth/login";
  return <p>Redirecting...</p>;
}

export default App;
