!function(){"use strict";var s=window.location,f=window.document,d=f.currentScript,v=d.getAttribute("data-api")||new URL(d.src).origin+"/api/event";function g(e){console.warn("Ignoring Event: "+e)}function e(e,t){try{if("true"===window.localStorage.vince_ignore)return g("localStorage flag")}catch(e){}var n=d&&d.getAttribute("data-include"),r=d&&d.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(o),i=r&&r.split(",").some(o);if(!a||i)return g("exclusion rule")}function o(e){return s.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=t&&t.u?t.u:s.href,c.d=d.getAttribute("data-domain"),c.r=f.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props);var u=d.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)}),p=c.p||{};u.forEach(function(e){var t=e.replace("event-",""),n=d.getAttribute(e);p[t]=p[t]||n}),c.p=p;var l=new XMLHttpRequest;l.open("POST",v,!0),l.setRequestHeader("Content-Type","text/plain"),l.send(JSON.stringify(c)),l.onreadystatechange=function(){4===l.readyState&&t&&t.callback&&t.callback()}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n=0;n<t.length;n++)e.apply(this,t[n]);var u=1;function r(e){if("auxclick"!==e.type||e.button===u){var t,n,r,a,i,o=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target);o&&o.href&&o.href.split("?")[0];if((i=o)&&i.href&&i.host&&i.host!==s.host)return t=e,r={name:"Outbound Link: Click",props:{url:(n=o).href}},a=!1,void(!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(t,n)?vince(r.name,{props:r.props}):(vince(r.name,{props:r.props,callback:c}),setTimeout(c,5e3),t.preventDefault()))}function c(){a||(a=!0,window.location=n.href)}}f.addEventListener("click",r),f.addEventListener("auxclick",r)}();