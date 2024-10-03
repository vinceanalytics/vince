!function(){"use strict";var o=window.location,l=window.document,u=l.currentScript,s=u.getAttribute("data-api")||new URL(u.src).origin+"/api/event";function c(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(o.hostname)||"file:"===o.protocol)return c("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return c(null,t);try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",t)}catch(e){}var r=u&&u.getAttribute("data-include"),a=u&&u.getAttribute("data-exclude");if("pageview"===e){r=!r||r.split(",").some(n),a=a&&a.split(",").some(n);if(!r||a)return c("exclusion rule",t)}function n(e){return o.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var r={},i=(r.n=e,r.u=t&&t.u?t.u:o.href,r.d=u.getAttribute("data-domain"),r.r=l.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),new XMLHttpRequest);i.open("POST",s,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(r)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function p(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var f=1;function a(e){var t,r;if("auxclick"!==e.type||e.button===f)return(t=function(e){for(;e&&(void 0===e.tagName||!p(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],!function e(t,r){if(!t||g<r)return!1;if(v(t))return!0;return e(t.parentNode,r+1)}(t,0)&&(r=t)&&r.href&&r.host&&r.host!==o.host?d(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):void 0}function d(e,t,r){var a,n=!1;function i(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(a={props:r.props},plausible(r.name,a)):(a={props:r.props,callback:i},plausible(r.name,a),setTimeout(i,5e3),e.preventDefault())}function m(e){var e=v(e)?e:e&&e.parentNode,t={name:null,props:{}},r=e&&e.classList;if(r)for(var a=0;a<r.length;a++){var n,i=r.item(a).match(/plausible-event-(.+)(=|--)(.+)/);i&&(n=i[1],i=i[3].replace(/\+/g," "),"name"==n.toLowerCase()?t.name=i:t.props[n]=i)}return t}l.addEventListener("click",a),l.addEventListener("auxclick",a);var g=3;function n(e){if("auxclick"!==e.type||e.button===f){for(var t,r,a,n,i=e.target,o=0;o<=g&&i;o++){if((a=i)&&a.tagName&&"form"===a.tagName.toLowerCase())return;p(i)&&(t=i),v(i)&&(r=i),i=i.parentNode}r&&(n=m(r),t?(n.props.url=t.href,d(e,t,n)):((e={}).props=n.props,plausible(n.name,e)))}}function v(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}l.addEventListener("submit",function(e){var t,r=e.target,a=m(r);function n(){t||(t=!0,r.submit())}a.name&&(e.preventDefault(),t=!1,setTimeout(n,5e3),e={props:a.props,callback:n},plausible(a.name,e))}),l.addEventListener("click",n),l.addEventListener("auxclick",n)}();