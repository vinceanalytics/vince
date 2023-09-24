import { Autocomplete } from "@primer/react"
import { useSites } from "../../providers"
import { useCallback } from "react"



export type SelectSiteProp = {
    selectSite: (site: string) => void,
}

export const SelectSite = ({ selectSite }: SelectSiteProp) => {
    const { sites } = useSites()
    const items = sites.map(({ domain }, id) => ({ id, text: domain }))
    return (
        <Autocomplete>
            <Autocomplete.Input block />
            <Autocomplete.Overlay>
                <Autocomplete.Menu
                    items={items}
                    selectedItemIds={[]}
                    selectionVariant="single"
                    onSelectedChange={e => {
                        const ls = e as { id: number, text: string }[]
                        selectSite(ls[0].text)
                    }}
                />
            </Autocomplete.Overlay>
        </Autocomplete>
    )
}