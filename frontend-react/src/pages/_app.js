import "@/styles/globals.css";
import Sidebar from "./components/Sidebar";

export default function App({ Component, pageProps }) {
  return (
    <div className="h-screen w-full flex">
      <Sidebar /> 
      <div className="w-10/12 h-full"> {/* div container for the content */}
        <Component {...pageProps} />
      </div>
    </div>
  )
}
