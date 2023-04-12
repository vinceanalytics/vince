!function(){"use strict";var e,t,n,r=window.location,i=window.document,o=i.getElementById("vince"),c=o.getAttribute("data-api")||(e=o.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function s(e){console.warn("Ignoring Event: "+e)}function a(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(r.hostname)||"file:"===r.protocol)return s("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return s("localStorage flag")}catch(e){}var n={};n.n=e,n.u=r.href,n.d=o.getAttribute("data-domain"),n.r=i.referrer||null,n.w=window.innerWidth,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props);var a=new XMLHttpRequest;a.open("POST",c,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(n)),a.onreadystatechange=function(){4===a.readyState&&t&&t.callback&&t.callback()}}}var p=window.vince&&window.vince.q||[];window.vince=a;for(var u,l=0;l<p.length;l++)a.apply(this,p[l]);function f(){u!==r.pathname&&(u=r.pathname,a("pageview"))}var v,d=window.history;function m(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}d.pushState&&(v=d.pushState,d.pushState=function(){v.apply(this,arguments),f()},window.addEventListener("popstate",f)),"prerender"===i.visibilityState?i.addEventListener("visibilitychange",function(){u||"visible"!==i.visibilityState||f()}):f();var w=1;function g(e){var t;"auxclick"===e.type&&e.button!==w||((t=function(e){for(;e&&(void 0===e.tagName||!m(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],function e(t,n){if(!t||b<n)return!1;if(k(t))return!0;return e(t.parentNode,n+1)}(t,0))}function h(e,t,n){var a=!1;function r(){a||(a=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),a=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&a}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:r}),setTimeout(r,5e3),e.preventDefault())}function y(e){var t=k(e)?e:e&&e.parentNode,n={name:null,props:{}},a=t&&t.classList;if(!a)return n;for(var r=0;r<a.length;r++){var i,o,c=a.item(r).match(/vince-event-(.+)=(.+)/);c&&(i=c[1],o=c[2].replace(/\+/g," "),"name"===i.toLowerCase()?n.name=o:n.props[i]=o)}return n}i.addEventListener("click",g),i.addEventListener("auxclick",g);var b=3;function L(e){if("auxclick"!==e.type||e.button===w){for(var t,n,a,r,i=e.target,o=0;o<=b&&i;o++){if((a=i)&&a.tagName&&"form"===a.tagName.toLowerCase())return;m(i)&&(t=i),k(i)&&(n=i),i=i.parentNode}n&&(r=y(n),t?(r.props.url=t.href,h(e,t,r)):vince(r.name,{props:r.props}))}}function k(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/vince-event-name=(.+)/))return 1}i.addEventListener("submit",function(e){var t,n=e.target,a=y(n);function r(){t||(t=!0,n.submit())}a.name&&(e.preventDefault(),t=!1,setTimeout(r,5e3),vince(a.name,{props:a.props,callback:r}))}),i.addEventListener("click",L),i.addEventListener("auxclick",L)}();