!function(){"use strict";var o=window.location,l=window.document,p=l.getElementById("plausible"),s=p.getAttribute("data-api")||(i=(i=p).src.split("/"),n=i[0],i=i[2],n+"//"+i+"/api/event");function u(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(o.hostname)||"file:"===o.protocol)return u("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return u(null,t);try{if("true"===window.localStorage.plausible_ignore)return u("localStorage flag",t)}catch(e){}var a=p&&p.getAttribute("data-include"),r=p&&p.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(n),r=r&&r.split(",").some(n);if(!a||r)return u("exclusion rule",t)}function n(e){return o.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},i=(a.n=e,a.u=t&&t.u?t.u:o.href,a.d=p.getAttribute("data-domain"),a.r=l.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),new XMLHttpRequest);i.open("POST",s,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(a)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a=0;a<t.length;a++)e.apply(this,t[a]);function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var f=1;function r(e){var t,a,r,n;if("auxclick"!==e.type||e.button===f)return t=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target),a=t&&t.href&&t.href.split("?")[0],!function e(t,a){if(!t||w<a)return!1;if(h(t))return!0;return e(t.parentNode,a+1)}(t,0)&&(r=a)&&(n=r.split(".").pop(),g.some(function(e){return e===n}))?d(e,t,{name:"File Download",props:{url:a}}):void 0}function d(e,t,a){var r,n=!1;function i(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(r={props:a.props},plausible(a.name,r)):(r={props:a.props,callback:i},plausible(a.name,r),setTimeout(i,5e3),e.preventDefault())}l.addEventListener("click",r),l.addEventListener("auxclick",r);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],i=p.getAttribute("file-types"),m=p.getAttribute("add-file-types"),g=i&&i.split(",")||m&&m.split(",").concat(n)||n;function v(e){var e=h(e)?e:e&&e.parentNode,t={name:null,props:{}},a=e&&e.classList;if(a)for(var r=0;r<a.length;r++){var n,i=a.item(r).match(/plausible-event-(.+)(=|--)(.+)/);i&&(n=i[1],i=i[3].replace(/\+/g," "),"name"==n.toLowerCase()?t.name=i:t.props[n]=i)}return t}var w=3;function b(e){if("auxclick"!==e.type||e.button===f){for(var t,a,r,n,i=e.target,o=0;o<=w&&i;o++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;c(i)&&(t=i),h(i)&&(a=i),i=i.parentNode}a&&(n=v(a),t?(n.props.url=t.href,d(e,t,n)):((e={}).props=n.props,plausible(n.name,e)))}}function h(e){var t=e&&e.classList;if(t)for(var a=0;a<t.length;a++)if(t.item(a).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}l.addEventListener("submit",function(e){var t,a=e.target,r=v(a);function n(){t||(t=!0,a.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(n,5e3),e={props:r.props,callback:n},plausible(r.name,e))}),l.addEventListener("click",b),l.addEventListener("auxclick",b)}();