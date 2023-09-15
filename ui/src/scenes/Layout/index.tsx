import { useCallback, useState } from "react"
import { VinceProvider, EditorProvider, SitesProvider, QueryProvider, TokenSourceProvider } from "../../providers";
import Footer from "../Footer";
import { Sidebar } from "../Sidebar"
import Editor from "../Editor"
import { Box, Portal, registerPortalRoot } from "@primer/react";
import { PageHeader } from '@primer/react/drafts'
import { useLocalStorage, StoreKey, SettingsType } from "../../providers/LocalStorageProvider"
import { Splitter, } from "../../components"
import Sites from "../Sites";
import { Auth } from "../Auth";
import { Result } from "../Result";
import { Settings } from "../Settings";



registerPortalRoot(document.getElementById("console")!, "console")
registerPortalRoot(document.getElementById("settings")!, "settings")
registerPortalRoot(document.getElementById("sidebar")!, "sidebar")
registerPortalRoot(document.getElementById("footer")!, "footer")

const Layout = () => {
    const [activePane, setActivePane] = useState<string>("console")

    const { editorSplitterBasis, resultsSplitterBasis, updateSettings } =
        useLocalStorage()
    const handleEditorSplitterChange = useCallback((value: SettingsType) => {
        updateSettings(StoreKey.EDITOR_SPLITTER_BASIS, value)
    }, [])

    const handleResultsSplitterChange = useCallback((value: SettingsType) => {
        updateSettings(StoreKey.RESULTS_SPLITTER_BASIS, value)
    }, [])

    const paneChange = (pane: string) => {
        setActivePane(pane)
    }

    return (
        <TokenSourceProvider>
            <VinceProvider>
                <SitesProvider>
                    <Auth>
                        <Sidebar onPanelChange={paneChange} />
                        <Footer />
                        <Portal containerName="console">
                            <Box
                                display={activePane === "console" ? "flex" : "none"}
                                flex={1}
                                flexDirection={"column"}
                                minHeight={"100%"}
                            >
                                <EditorProvider>
                                    <QueryProvider>
                                        <Splitter
                                            direction="vertical"
                                            fallback={editorSplitterBasis}
                                            min={100}
                                            onChange={handleEditorSplitterChange}
                                        >
                                            <Box
                                                position={"relative"}
                                                overflow={"hidden"}
                                            >
                                                <Splitter
                                                    direction="horizontal"
                                                    fallback={resultsSplitterBasis}
                                                    max={500}
                                                    onChange={handleResultsSplitterChange}
                                                >
                                                    <Box >
                                                        <Sites />
                                                    </Box>
                                                    <Editor />
                                                </Splitter>
                                            </Box>
                                            <Result />
                                        </Splitter>
                                    </QueryProvider>
                                </EditorProvider>
                            </Box>
                        </Portal>
                        <Portal containerName="settings">
                            <Settings activePane={activePane} />
                        </Portal>
                    </Auth>
                </SitesProvider>
            </VinceProvider >
        </TokenSourceProvider>
    )
}
export default Layout