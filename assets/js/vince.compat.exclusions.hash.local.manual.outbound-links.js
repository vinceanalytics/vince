!function(){"use strict";var e,t,n,l=window.location,p=window.document,s=p.getElementById("vince"),d=s.getAttribute("data-api")||(e=s.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function f(e){console.warn("Ignoring Event: "+e)}function r(e,t){try{if("true"===window.localStorage.vince_ignore)return f("localStorage flag")}catch(e){}var n=s&&s.getAttribute("data-include"),r=s&&s.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(o),i=r&&r.split(",").some(o);if(!a||i)return f("exclusion rule")}function o(e){var t=l.pathname;return(t+=l.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=t&&t.u?t.u:l.href,c.d=s.getAttribute("data-domain"),c.r=p.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props),c.h=1;var u=new XMLHttpRequest;u.open("POST",d,!0),u.setRequestHeader("Content-Type","text/plain"),u.send(JSON.stringify(c)),u.onreadystatechange=function(){4===u.readyState&&t&&t.callback&&t.callback()}}var a=window.vince&&window.vince.q||[];window.vince=r;for(var i=0;i<a.length;i++)r.apply(this,a[i]);var u=1;function o(e){if("auxclick"!==e.type||e.button===u){var t,n,r,a,i,o=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target);o&&o.href&&o.href.split("?")[0];if((i=o)&&i.href&&i.host&&i.host!==l.host)return t=e,r={name:"Outbound Link: Click",props:{url:(n=o).href}},a=!1,void(!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(t,n)?vince(r.name,{props:r.props}):(vince(r.name,{props:r.props,callback:c}),setTimeout(c,5e3),t.preventDefault()))}function c(){a||(a=!0,window.location=n.href)}}p.addEventListener("click",o),p.addEventListener("auxclick",o)}();