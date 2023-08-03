import { VinceProvider } from "../../providers";
import Footer from "../Footer";
import Sidebar from "../Sidebar"

const Layout = () => {
    return (
        <VinceProvider>
            <Sidebar />
            <Footer />
        </VinceProvider>
    )
}
export default Layout