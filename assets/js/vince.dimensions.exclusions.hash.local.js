!function(){"use strict";var d=window.location,p=window.document,v=p.currentScript,g=v.getAttribute("data-api")||new URL(v.src).origin+"/api/event";function w(e){console.warn("Ignoring Event: "+e)}function e(e,t){try{if("true"===window.localStorage.vince_ignore)return w("localStorage flag")}catch(e){}var n=v&&v.getAttribute("data-include"),i=v&&v.getAttribute("data-exclude");if("pageview"===e){var r=!n||n&&n.split(",").some(o),a=i&&i.split(",").some(o);if(!r||a)return w("exclusion rule")}function o(e){var t=d.pathname;return(t+=d.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=d.href,c.d=v.getAttribute("data-domain"),c.r=p.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props);var u=v.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)}),s=c.p||{};u.forEach(function(e){var t=e.replace("event-",""),n=v.getAttribute(e);s[t]=s[t]||n}),c.p=s,c.h=1;var l=new XMLHttpRequest;l.open("POST",g,!0),l.setRequestHeader("Content-Type","text/plain"),l.send(JSON.stringify(c)),l.onreadystatechange=function(){4===l.readyState&&t&&t.callback&&t.callback()}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n,i=0;i<t.length;i++)e.apply(this,t[i]);function r(){n=d.pathname,e("pageview")}window.addEventListener("hashchange",r),"prerender"===p.visibilityState?p.addEventListener("visibilitychange",function(){n||"visible"!==p.visibilityState||r()}):r()}();