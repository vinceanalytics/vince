!function(){"use strict";var e,t,n,o=window.location,c=window.document,s=c.getElementById("vince"),p=s.getAttribute("data-api")||(e=s.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function u(e){console.warn("Ignoring Event: "+e)}function r(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(o.hostname)||"file:"===o.protocol)return u("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return u("localStorage flag")}catch(e){}var n={};n.n=e,n.u=o.href,n.d=s.getAttribute("data-domain"),n.r=c.referrer||null,n.w=window.innerWidth,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props);var r=s.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)}),a=n.p||{};r.forEach(function(e){var t=e.replace("event-",""),n=s.getAttribute(e);a[t]=a[t]||n}),n.p=a;var i=new XMLHttpRequest;i.open("POST",p,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(n)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback()}}}var a=window.vince&&window.vince.q||[];window.vince=r;for(var i,f=0;f<a.length;f++)r.apply(this,a[f]);function l(){i!==o.pathname&&(i=o.pathname,r("pageview"))}var v,d=window.history;function m(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}d.pushState&&(v=d.pushState,d.pushState=function(){v.apply(this,arguments),l()},window.addEventListener("popstate",l)),"prerender"===c.visibilityState?c.addEventListener("visibilitychange",function(){i||"visible"!==c.visibilityState||l()}):l();var h=1;function w(e){if("auxclick"!==e.type||e.button===h){var t,n=function(e){for(;e&&(void 0===e.tagName||!m(e)||!e.href);)e=e.parentNode;return e}(e.target);n&&n.href&&n.href.split("?")[0];if(!function e(t,n){if(!t||b<n)return!1;if(k(t))return!0;return e(t.parentNode,n+1)}(n,0))return(t=n)&&t.href&&t.host&&t.host!==o.host?g(e,n,{name:"Outbound Link: Click",props:{url:n.href}}):void 0}}function g(e,t,n){var r=!1;function a(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:a}),setTimeout(a,5e3),e.preventDefault())}function y(e){var t=k(e)?e:e&&e.parentNode,n={name:null,props:{}},r=t&&t.classList;if(!r)return n;for(var a=0;a<r.length;a++){var i,o,c=r.item(a).match(/vince-event-(.+)=(.+)/);c&&(i=c[1],o=c[2].replace(/\+/g," "),"name"===i.toLowerCase()?n.name=o:n.props[i]=o)}return n}c.addEventListener("click",w),c.addEventListener("auxclick",w);var b=3;function L(e){if("auxclick"!==e.type||e.button===h){for(var t,n,r,a,i=e.target,o=0;o<=b&&i;o++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;m(i)&&(t=i),k(i)&&(n=i),i=i.parentNode}n&&(a=y(n),t?(a.props.url=t.href,g(e,t,a)):vince(a.name,{props:a.props}))}}function k(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/vince-event-name=(.+)/))return 1}c.addEventListener("submit",function(e){var t,n=e.target,r=y(n);function a(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),vince(r.name,{props:r.props,callback:a}))}),c.addEventListener("click",L),c.addEventListener("auxclick",L)}();