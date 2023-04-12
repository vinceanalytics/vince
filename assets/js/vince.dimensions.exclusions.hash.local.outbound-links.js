!function(){"use strict";var p=window.location,d=window.document,f=d.currentScript,v=f.getAttribute("data-api")||new URL(f.src).origin+"/api/event";function g(e){console.warn("Ignoring Event: "+e)}function e(e,t){try{if("true"===window.localStorage.vince_ignore)return g("localStorage flag")}catch(e){}var n=f&&f.getAttribute("data-include"),r=f&&f.getAttribute("data-exclude");if("pageview"===e){var i=!n||n&&n.split(",").some(o),a=r&&r.split(",").some(o);if(!i||a)return g("exclusion rule")}function o(e){var t=p.pathname;return(t+=p.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=p.href,c.d=f.getAttribute("data-domain"),c.r=d.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props);var u=f.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)}),s=c.p||{};u.forEach(function(e){var t=e.replace("event-",""),n=f.getAttribute(e);s[t]=s[t]||n}),c.p=s,c.h=1;var l=new XMLHttpRequest;l.open("POST",v,!0),l.setRequestHeader("Content-Type","text/plain"),l.send(JSON.stringify(c)),l.onreadystatechange=function(){4===l.readyState&&t&&t.callback&&t.callback()}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n,r=0;r<t.length;r++)e.apply(this,t[r]);function i(){n=p.pathname,e("pageview")}window.addEventListener("hashchange",i),"prerender"===d.visibilityState?d.addEventListener("visibilitychange",function(){n||"visible"!==d.visibilityState||i()}):i();var u=1;function a(e){if("auxclick"!==e.type||e.button===u){var t,n,r,i,a,o=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target);o&&o.href&&o.href.split("?")[0];if((a=o)&&a.href&&a.host&&a.host!==p.host)return t=e,r={name:"Outbound Link: Click",props:{url:(n=o).href}},i=!1,void(!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(t,n)?vince(r.name,{props:r.props}):(vince(r.name,{props:r.props,callback:c}),setTimeout(c,5e3),t.preventDefault()))}function c(){i||(i=!0,window.location=n.href)}}d.addEventListener("click",a),d.addEventListener("auxclick",a)}();