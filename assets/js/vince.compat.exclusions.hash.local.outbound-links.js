!function(){"use strict";var e,t,n,s=window.location,p=window.document,u=p.getElementById("vince"),d=u.getAttribute("data-api")||(e=u.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function v(e){console.warn("Ignoring Event: "+e)}function i(e,t){try{if("true"===window.localStorage.vince_ignore)return v("localStorage flag")}catch(e){}var n=u&&u.getAttribute("data-include"),i=u&&u.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(o),r=i&&i.split(",").some(o);if(!a||r)return v("exclusion rule")}function o(e){var t=s.pathname;return(t+=s.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=s.href,c.d=u.getAttribute("data-domain"),c.r=p.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props),c.h=1;var l=new XMLHttpRequest;l.open("POST",d,!0),l.setRequestHeader("Content-Type","text/plain"),l.send(JSON.stringify(c)),l.onreadystatechange=function(){4===l.readyState&&t&&t.callback&&t.callback()}}var a=window.vince&&window.vince.q||[];window.vince=i;for(var r,o=0;o<a.length;o++)i.apply(this,a[o]);function c(){r=s.pathname,i("pageview")}window.addEventListener("hashchange",c),"prerender"===p.visibilityState?p.addEventListener("visibilitychange",function(){r||"visible"!==p.visibilityState||c()}):c();var l=1;function f(e){if("auxclick"!==e.type||e.button===l){var t,n,i,a,r,o=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target);o&&o.href&&o.href.split("?")[0];if((r=o)&&r.href&&r.host&&r.host!==s.host)return t=e,i={name:"Outbound Link: Click",props:{url:(n=o).href}},a=!1,void(!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),i=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&i}}(t,n)?vince(i.name,{props:i.props}):(vince(i.name,{props:i.props,callback:c}),setTimeout(c,5e3),t.preventDefault()))}function c(){a||(a=!0,window.location=n.href)}}p.addEventListener("click",f),p.addEventListener("auxclick",f)}();