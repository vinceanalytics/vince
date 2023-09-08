import { useCallback, useState } from "react"
import { VinceProvider, EditorProvider, SitesProvider, QueryProvider } from "../../providers";
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
                            height={"100vh"}
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
                        <Box
                            display={activePane === "settings" ? "flex" : "none"}
                            flex={1}
                            flexDirection={"column"}
                            maxHeight={"100%"}
                        >
                            <PageHeader>
                                <PageHeader.TitleArea>
                                    <PageHeader.Title>Settings</PageHeader.Title>
                                </PageHeader.TitleArea>
                            </PageHeader>
                        </Box>
                    </Portal>
                </Auth>
            </SitesProvider>
        </VinceProvider >
    )
}
export default Layout