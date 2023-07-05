
import uPlot from 'uplot';

function plotSeries(id: string, label: string, x: number[], y: number[]) {
    let el = document.getElementById(id)!
    const container = el.parentElement!;
    let p = new uPlot({ width: container.scrollWidth!, height: 240, series: [{}, { label: label, fill: "#ffe0d8" }], ms: 1 }, [x, y], el!);;
    const resize = new ResizeObserver(() => {
        p.setSize({
            width: container.scrollWidth!,
            height: 240,
        })
    });
    resize.observe(container);
}

//@ts-ignore
window.plotSeries = plotSeries;