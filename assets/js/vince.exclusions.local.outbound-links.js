!function(){"use strict";var s=window.location,u=window.document,l=u.currentScript,d=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function f(e){console.warn("Ignoring Event: "+e)}function e(e,t){try{if("true"===window.localStorage.vince_ignore)return f("localStorage flag")}catch(e){}var n=l&&l.getAttribute("data-include"),i=l&&l.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(o),r=i&&i.split(",").some(o);if(!a||r)return f("exclusion rule")}function o(e){return s.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=s.href,c.d=l.getAttribute("data-domain"),c.r=u.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props);var p=new XMLHttpRequest;p.open("POST",d,!0),p.setRequestHeader("Content-Type","text/plain"),p.send(JSON.stringify(c)),p.onreadystatechange=function(){4===p.readyState&&t&&t.callback&&t.callback()}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n,i=0;i<t.length;i++)e.apply(this,t[i]);function a(){n!==s.pathname&&(n=s.pathname,e("pageview"))}var r,o=window.history;o.pushState&&(r=o.pushState,o.pushState=function(){r.apply(this,arguments),a()},window.addEventListener("popstate",a)),"prerender"===u.visibilityState?u.addEventListener("visibilitychange",function(){n||"visible"!==u.visibilityState||a()}):a();var p=1;function c(e){if("auxclick"!==e.type||e.button===p){var t,n,i,a,r,o=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target);o&&o.href&&o.href.split("?")[0];if((r=o)&&r.href&&r.host&&r.host!==s.host)return t=e,i={name:"Outbound Link: Click",props:{url:(n=o).href}},a=!1,void(!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),i=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&i}}(t,n)?vince(i.name,{props:i.props}):(vince(i.name,{props:i.props,callback:c}),setTimeout(c,5e3),t.preventDefault()))}function c(){a||(a=!0,window.location=n.href)}}u.addEventListener("click",c),u.addEventListener("auxclick",c)}();