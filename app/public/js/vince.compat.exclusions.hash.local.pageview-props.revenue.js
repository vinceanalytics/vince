!function(){"use strict";var e,t,s=window.location,u=window.document,o=u.getElementById("plausible"),c=o.getAttribute("data-api")||(e=(e=o).src.split("/"),t=e[0],e=e[2],t+"//"+e+"/api/event");function p(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function a(e,t){try{if("true"===window.localStorage.plausible_ignore)return p("localStorage flag",t)}catch(e){}var a=o&&o.getAttribute("data-include"),n=o&&o.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(i),n=n&&n.split(",").some(i);if(!a||n)return p("exclusion rule",t)}function i(e){var t=s.pathname;return(t+=s.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},n=(a.n=e,a.u=s.href,a.d=o.getAttribute("data-domain"),a.r=u.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),o.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),r=a.p||{},l=(n.forEach(function(e){var t=e.replace("event-",""),e=o.getAttribute(e);r[t]=r[t]||e}),a.p=r,a.h=1,new XMLHttpRequest);l.open("POST",c,!0),l.setRequestHeader("Content-Type","text/plain"),l.send(JSON.stringify(a)),l.onreadystatechange=function(){4===l.readyState&&t&&t.callback&&t.callback({status:l.status})}}var n=window.plausible&&window.plausible.q||[];window.plausible=a;for(var i,r=0;r<n.length;r++)a.apply(this,n[r]);function l(){i=s.pathname,a("pageview")}window.addEventListener("hashchange",l),"prerender"===u.visibilityState?u.addEventListener("visibilitychange",function(){i||"visible"!==u.visibilityState||l()}):l()}();