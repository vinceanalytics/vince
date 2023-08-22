import { createContext, PropsWithChildren, useState, useContext, useEffect } from "react"
import { Site, SiteList } from "../../vince"
import { useVince } from "../VinceProvider"



type ContextProps = {
    sites: Site[]
    refresh: () => void
}

const defaultValues = {
    sites: [],
    refresh: () => undefined,
}

const SitesContext = createContext<ContextProps>(defaultValues)


export const SitesProvider = ({ children }: PropsWithChildren<{}>) => {
    const [sites, setSites] = useState<Site[]>([])
    const { vince } = useVince()
    const refresh = () => {
        vince.sites().then((result) => {
            const { list } = result as SiteList;
            setSites(list)
        })
            .catch((e) => {
                console.log(e)
            })
    }
    useEffect(() => {
        if (vince.authenticated()) {
            refresh()
        }
    }, [vince])

    return (
        <SitesContext.Provider
            value={{
                sites,
                refresh,
            }}
        >
            {children}
        </SitesContext.Provider>
    )
}

export const useSites = () => {
    return useContext(SitesContext)
}