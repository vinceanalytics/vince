!function(){"use strict";var e,t,n,p=window.location,d=window.document,v=d.getElementById("vince"),f=v.getAttribute("data-api")||(e=v.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function g(e){console.warn("Ignoring Event: "+e)}function i(e,t){try{if("true"===window.localStorage.vince_ignore)return g("localStorage flag")}catch(e){}var n=v&&v.getAttribute("data-include"),i=v&&v.getAttribute("data-exclude");if("pageview"===e){var r=!n||n&&n.split(",").some(o),a=i&&i.split(",").some(o);if(!r||a)return g("exclusion rule")}function o(e){var t=p.pathname;return(t+=p.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=p.href,c.d=v.getAttribute("data-domain"),c.r=d.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props);var s=v.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)}),u=c.p||{};s.forEach(function(e){var t=e.replace("event-",""),n=v.getAttribute(e);u[t]=u[t]||n}),c.p=u,c.h=1;var l=new XMLHttpRequest;l.open("POST",f,!0),l.setRequestHeader("Content-Type","text/plain"),l.send(JSON.stringify(c)),l.onreadystatechange=function(){4===l.readyState&&t&&t.callback&&t.callback()}}var r=window.vince&&window.vince.q||[];window.vince=i;for(var a,o=0;o<r.length;o++)i.apply(this,r[o]);function c(){a=p.pathname,i("pageview")}window.addEventListener("hashchange",c),"prerender"===d.visibilityState?d.addEventListener("visibilitychange",function(){a||"visible"!==d.visibilityState||c()}):c();var s=1;function u(e){if("auxclick"!==e.type||e.button===s){var t,n,i,r,a,o=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target);o&&o.href&&o.href.split("?")[0];if((a=o)&&a.href&&a.host&&a.host!==p.host)return t=e,i={name:"Outbound Link: Click",props:{url:(n=o).href}},r=!1,void(!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),i=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&i}}(t,n)?vince(i.name,{props:i.props}):(vince(i.name,{props:i.props,callback:c}),setTimeout(c,5e3),t.preventDefault()))}function c(){r||(r=!0,window.location=n.href)}}d.addEventListener("click",u),d.addEventListener("auxclick",u)}();