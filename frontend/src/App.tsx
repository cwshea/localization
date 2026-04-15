import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import TranslationList from "./components/TranslationList";
import TranslationForm from "./components/TranslationForm";
import TranslationDetail from "./components/TranslationDetail";
import "./App.css";

function App() {
  return (
    <BrowserRouter>
      <div className="app">
        <header className="app-header">
          <Link to="/" className="app-title">
            Localization
          </Link>
        </header>
        <main className="app-main">
          <Routes>
            <Route path="/" element={<TranslationList />} />
            <Route path="/new" element={<TranslationForm />} />
            <Route path="/source/:id" element={<TranslationDetail />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  );
}

export default App;
