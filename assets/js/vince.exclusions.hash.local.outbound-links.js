!function(){"use strict";var l=window.location,u=window.document,p=u.currentScript,d=p.getAttribute("data-api")||new URL(p.src).origin+"/api/event";function f(e){console.warn("Ignoring Event: "+e)}function e(e,t){try{if("true"===window.localStorage.vince_ignore)return f("localStorage flag")}catch(e){}var n=p&&p.getAttribute("data-include"),i=p&&p.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(o),r=i&&i.split(",").some(o);if(!a||r)return f("exclusion rule")}function o(e){var t=l.pathname;return(t+=l.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=l.href,c.d=p.getAttribute("data-domain"),c.r=u.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props),c.h=1;var s=new XMLHttpRequest;s.open("POST",d,!0),s.setRequestHeader("Content-Type","text/plain"),s.send(JSON.stringify(c)),s.onreadystatechange=function(){4===s.readyState&&t&&t.callback&&t.callback()}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n,i=0;i<t.length;i++)e.apply(this,t[i]);function a(){n=l.pathname,e("pageview")}window.addEventListener("hashchange",a),"prerender"===u.visibilityState?u.addEventListener("visibilitychange",function(){n||"visible"!==u.visibilityState||a()}):a();var s=1;function r(e){if("auxclick"!==e.type||e.button===s){var t,n,i,a,r,o=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target);o&&o.href&&o.href.split("?")[0];if((r=o)&&r.href&&r.host&&r.host!==l.host)return t=e,i={name:"Outbound Link: Click",props:{url:(n=o).href}},a=!1,void(!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),i=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&i}}(t,n)?vince(i.name,{props:i.props}):(vince(i.name,{props:i.props,callback:c}),setTimeout(c,5e3),t.preventDefault()))}function c(){a||(a=!0,window.location=n.href)}}u.addEventListener("click",r),u.addEventListener("auxclick",r)}();