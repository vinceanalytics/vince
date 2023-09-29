import { createContext, PropsWithChildren, useState, useContext, useEffect } from "react"
import { Site } from "../../vince"
import { useVince } from "../VinceProvider"
import { useLocalStorage, StoreKey } from "../LocalStorageProvider"
import { RpcError } from "@protobuf-ts/runtime-rpc"



type ContextProps = {
    sites: Site[]
    refresh: () => void
    selectedSite: string,
    selectSite: (site: string) => void,
}

const defaultValues = {
    sites: [],
    refresh: () => undefined,
    selectedSite: "(no site)",
    selectSite: () => undefined,
}

const SitesContext = createContext<ContextProps>(defaultValues)


export const SitesProvider = ({ children }: PropsWithChildren<{}>) => {
    const [sites, setSites] = useState<Site[]>([])
    const { sitesClient } = useVince()
    const [selectedSite, setSelectedSite] = useState<string>("(no site)")
    const { updateSettings } = useLocalStorage()
    const refresh = () => {
        sitesClient?.listSites({}).then((result) => {
            setSites(result.response.list)
        })
            .catch((e: RpcError) => {
                if (e.code === "UNAUTHENTICATED") {
                    // TODO: use the refresh_token to request a new token
                    updateSettings(StoreKey.AUTH_PAYLOAD, "")
                }
            })
    }
    const selectSite = (id: string) => setSelectedSite(id)

    useEffect(() => {
        if (sitesClient) {
            refresh()
        }
    }, [sitesClient])

    return (
        <SitesContext.Provider
            value={{
                sites,
                refresh,
                selectedSite, selectSite,
            }}
        >
            {children}
        </SitesContext.Provider>
    )
}

export const useSites = () => {
    return useContext(SitesContext)
}