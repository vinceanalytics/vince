import { createContext, PropsWithChildren, useState, useContext } from "react"
import { Site } from "../../vince"



type ContextProps = {
    sites: Site[] | null
    updateSites: (sites: Site[]) => void
}

const defaultValues = {
    sites: [],
    updateSites: (sites: Site[]) => undefined,
}

const SitesContext = createContext<ContextProps>(defaultValues)


export const SitesProvider = ({ children }: PropsWithChildren<{}>) => {
    const [sites, setSites] = useState<Site[]>([])
    const updateSites = (site: Site[]) => {
        setSites(site)
    }

    return (
        <SitesContext.Provider
            value={{
                sites,
                updateSites,
            }}
        >
            {children}
        </SitesContext.Provider>
    )
}

export const useSites = () => {
    return useContext(SitesContext)
}