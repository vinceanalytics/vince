!function(){"use strict";var i=window.location,a=window.document,o=a.currentScript,p=o.getAttribute("data-api")||new URL(o.src).origin+"/api/event";function c(t){console.warn("Ignoring Event: "+t)}function t(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(i.hostname)||"file:"===i.protocol)return c("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return c("localStorage flag")}catch(t){}var n={};n.n=t,n.u=i.href,n.d=o.getAttribute("data-domain"),n.r=a.referrer||null,n.w=window.innerWidth,e&&e.meta&&(n.m=JSON.stringify(e.meta)),e&&e.props&&(n.p=e.props);var r=new XMLHttpRequest;r.open("POST",p,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(n)),r.onreadystatechange=function(){4===r.readyState&&e&&e.callback&&e.callback()}}}var e=window.vince&&window.vince.q||[];window.vince=t;for(var n,r=0;r<e.length;r++)t.apply(this,e[r]);function s(){n!==i.pathname&&(n=i.pathname,t("pageview"))}var u,l=window.history;function f(t){return t&&t.tagName&&"a"===t.tagName.toLowerCase()}l.pushState&&(u=l.pushState,l.pushState=function(){u.apply(this,arguments),s()},window.addEventListener("popstate",s)),"prerender"===a.visibilityState?a.addEventListener("visibilitychange",function(){n||"visible"!==a.visibilityState||s()}):s();var v=1;function d(t){if("auxclick"!==t.type||t.button===v){var e,n=function(t){for(;t&&(void 0===t.tagName||!f(t)||!t.href);)t=t.parentNode;return t}(t.target),r=n&&n.href&&n.href.split("?")[0];if(!function t(e,n){if(!e||k<n)return!1;if(S(e))return!0;return t(e.parentNode,n+1)}(n,0))return(e=n)&&e.href&&e.host&&e.host!==i.host?m(t,n,{name:"Outbound Link: Click",props:{url:n.href}}):function(t){if(!t)return!1;var e=t.split(".").pop();return y.some(function(t){return t===e})}(r)?m(t,n,{name:"File Download",props:{url:r}}):void 0}}function m(t,e,n){var r=!1;function i(){r||(r=!0,window.location=e.href)}!function(t,e){if(!t.defaultPrevented){var n=!e.target||e.target.match(/^_(self|parent|top)$/i),r=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type;return n&&r}}(t,e)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:i}),setTimeout(i,5e3),t.preventDefault())}a.addEventListener("click",d),a.addEventListener("auxclick",d);var w=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],g=o.getAttribute("file-types"),h=o.getAttribute("add-file-types"),y=g&&g.split(",")||h&&h.split(",").concat(w)||w;function b(t){var e=S(t)?t:t&&t.parentNode,n={name:null,props:{}},r=e&&e.classList;if(!r)return n;for(var i=0;i<r.length;i++){var a,o,p=r.item(i).match(/vince-event-(.+)=(.+)/);p&&(a=p[1],o=p[2].replace(/\+/g," "),"name"===a.toLowerCase()?n.name=o:n.props[a]=o)}return n}var k=3;function L(t){if("auxclick"!==t.type||t.button===v){for(var e,n,r,i,a=t.target,o=0;o<=k&&a;o++){if((r=a)&&r.tagName&&"form"===r.tagName.toLowerCase())return;f(a)&&(e=a),S(a)&&(n=a),a=a.parentNode}n&&(i=b(n),e?(i.props.url=e.href,m(t,e,i)):vince(i.name,{props:i.props}))}}function S(t){var e=t&&t.classList;if(e)for(var n=0;n<e.length;n++)if(e.item(n).match(/vince-event-name=(.+)/))return 1}a.addEventListener("submit",function(t){var e,n=t.target,r=b(n);function i(){e||(e=!0,n.submit())}r.name&&(t.preventDefault(),e=!1,setTimeout(i,5e3),vince(r.name,{props:r.props,callback:i}))}),a.addEventListener("click",L),a.addEventListener("auxclick",L)}();